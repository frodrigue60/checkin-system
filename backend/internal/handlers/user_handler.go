package handlers

import (
	"attendance-api/internal/config"
	"attendance-api/internal/models"
	"attendance-api/internal/services"
	"attendance-api/internal/utils"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type UserHandler struct {
	DB      *sqlx.DB
	Cfg     *config.Config
	Storage services.StorageService
}

func (h *UserHandler) GetAvatarUploadURL(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	fileName := c.Query("fileName", "avatar.jpg")

	uploadURL, _, key, err := h.Storage.GetAvatarUploadURL(c.Context(), userID, fileName)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Could not generate upload URL", err)
	}

	return c.JSON(fiber.Map{
		"uploadURL": uploadURL,
		"key":       key,
	})
}

func (h *UserHandler) ConfirmAvatarUpdate(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	var req struct {
		Key string `json:"key"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	_, err := h.DB.Exec("UPDATE users SET photo_url = $1, updated_at = NOW() WHERE id = $2", req.Key, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Could not update profile photo", err)
	}

	return c.JSON(fiber.Map{"message": "Photo updated successfully"})
}


func (h *UserHandler) GetEmployeeStats(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	// 1. Get Employee ID
	var employee models.Employee
	err := h.DB.Get(&employee, "SELECT * FROM employees WHERE user_id = $1", userID)
	if err != nil {
		// If not an employee, return empty stats instead of 404
		return c.JSON(fiber.Map{
			"stats": fiber.Map{
				"total_hours":     0,
				"total_earnings":  0,
				"incidents_count": 0,
			},
			"recent": []interface{}{},
		})
	}

	// 2. Calculate Month Stats
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var stats struct {
		TotalHours     float64 `db:"total_hours" json:"total_hours"`
		TotalEarnings  float64 `db:"total_earnings" json:"total_earnings"`
		IncidentsCount int     `db:"incidents_count" json:"incidents_count"`
	}

	query := `
		SELECT 
			COALESCE(SUM(net_hours_worked), 0) as total_hours,
			COALESCE(SUM(daily_earnings), 0) as total_earnings,
			(SELECT COUNT(*) FROM incidents i JOIN attendances a ON i.attendance_id = a.id WHERE a.employee_id = $1 AND a.check_in >= $2) as incidents_count
		FROM attendances 
		WHERE employee_id = $1 AND check_in >= $2
	`
	err = h.DB.Get(&stats, query, employee.ID, firstOfMonth)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	// 3. Get Recent Activity
	var entities []models.Attendance
	err = h.DB.Select(&entities, "SELECT * FROM attendances WHERE employee_id = $1 ORDER BY check_in DESC, created_at DESC LIMIT 5", employee.ID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	dtos := make([]models.AttendanceDTO, 0)
	for _, a := range entities {
		dtos = append(dtos, models.MapAttendanceToDTO(a, h.Cfg.R2PublicURL))
	}

	return c.JSON(fiber.Map{
		"stats":  stats,
		"recent": dtos,
	})
}

func (h *UserHandler) GetAttendanceHistory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	month := c.QueryInt("month", int(time.Now().Month()))
	year := c.QueryInt("year", time.Now().Year())

	// 1. Get Employee ID
	var employee models.Employee
	err := h.DB.Get(&employee, "SELECT * FROM employees WHERE user_id = $1", userID)
	if err != nil {
		// Return empty history instead of 404
		return c.JSON(fiber.Map{
			"stats": fiber.Map{
				"total_hours":     0,
				"attendance_rate": 0,
			},
			"history": []interface{}{},
		})
	}

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	// 2. Query Attendances with Work Center Name and Incidents
	type HistoryItem struct {
		ID             int              `db:"id" json:"id"`
		CheckIn        *time.Time       `db:"check_in" json:"check_in"`
		CheckOut       *time.Time       `db:"check_out" json:"check_out"`
		LunchStart     *time.Time       `db:"lunch_start" json:"lunch_start"`
		LunchEnd       *time.Time       `db:"lunch_end" json:"lunch_end"`
		WorkCenterName string           `db:"work_center_name" json:"work_center_name"`
		NetHoursWorked float64          `db:"net_hours_worked" json:"net_hours_worked"`
		IsAbsence      bool             `db:"is_absence" json:"is_absence"`
		AbsenceReason  *string          `db:"absence_reason" json:"absence_reason"`
		Incidents      []models.Incident `json:"incidents"`
	}

	var history []HistoryItem
	query := `
		SELECT a.id, a.check_in, a.check_out, a.lunch_start, a.lunch_end, a.net_hours_worked,
		       a.is_absence, a.absence_reason,
		       COALESCE(wc.name, 'N/A') as work_center_name
		FROM attendances a
		LEFT JOIN work_centers wc ON a.work_center_id = wc.id
		WHERE a.employee_id = $1 AND (a.check_in BETWEEN $2 AND $3 OR a.created_at BETWEEN $2 AND $3)
		ORDER BY a.created_at DESC
	`
	err = h.DB.Select(&history, query, employee.ID, startDate, endDate)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	// 2.1 Fetch All Incidents for the whole range in ONE query (Avoid N+1)
	attendanceIDs := make([]int, len(history))
	for i, item := range history {
		attendanceIDs[i] = item.ID
	}

	if len(attendanceIDs) > 0 {
		incQuery, incArgs, _ := sqlx.In(`
			SELECT i.*, 
			       (SELECT COUNT(*) FROM justifications j WHERE j.incident_id = i.id) > 0 as has_justification
			FROM incidents i 
			WHERE i.attendance_id IN (?)`, attendanceIDs)
		incQuery = h.DB.Rebind(incQuery)

		type IncidentWithFlag struct {
			models.Incident
			HasJustification bool `db:"has_justification" json:"has_justification"`
		}
		var allIncidents []IncidentWithFlag
		
		// Use context with timeout for safety
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()
		
		err = h.DB.SelectContext(ctx, &allIncidents, incQuery, incArgs...)
		if err == nil {
			// Map incidents to their respective attendance records
			incMap := make(map[int][]models.Incident)
			for _, r := range allIncidents {
				if r.HasJustification {
					r.Incident.Status = "pending_review"
				}
				incMap[r.AttendanceID] = append(incMap[r.AttendanceID], r.Incident)
			}

			for i := range history {
				if incidents, ok := incMap[history[i].ID]; ok {
					history[i].Incidents = incidents
				} else {
					history[i].Incidents = []models.Incident{}
				}
			}
		}
	}

	// 3. Stats for the month
	var totalHours float64
	for _, item := range history {
		totalHours += item.NetHoursWorked
	}

	// Simple attendance rate: days with logs / total days in month so far (or total passed days)
	daysInMonth := time.Now().Day()
	if time.Now().Month() != time.Month(month) || time.Now().Year() != year {
		daysInMonth = startDate.AddDate(0, 1, 0).AddDate(0, 0, -1).Day()
	}
	
	attendanceRate := 0.0
	if daysInMonth > 0 {
		attendanceRate = (float64(len(history)) / float64(daysInMonth)) * 100
	}

	return c.JSON(fiber.Map{
		"stats": fiber.Map{
			"total_hours":     totalHours,
			"attendance_rate": attendanceRate,
		},
		"history": history,
	})
}

func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	type ProfileData struct {
		ID                int        `db:"id" json:"id"`
		Name              string     `db:"name" json:"name"`
		Email             string     `db:"email" json:"email"`
		Phone             *string    `db:"phone" json:"phone"`
		PositionName      *string    `db:"position_name" json:"position_name"`
		WorkCenterName    *string    `db:"work_center_name" json:"work_center_name"`
		ShiftName         *string    `db:"shift_name" json:"shift_name"`
		ExpectedCheckIn   *string    `db:"expected_check_in" json:"expected_check_in"`
		ExpectedCheckOut  *string    `db:"expected_check_out" json:"expected_check_out"`
		EmployeeCreatedAt *time.Time `db:"employee_created_at" json:"employee_created_at"`
		IsActive          *bool      `db:"is_active" json:"is_active"`
		HourlyRate        *float64   `db:"hourly_rate" json:"hourly_rate"`
		PhotoURL          *string    `db:"photo_url" json:"photo_url"`
		IsEmployee        bool       `json:"is_employee"`
	}

	var profile ProfileData
	query := `
		SELECT u.id, u.name, u.email, u.phone, u.photo_url, p.name as position_name, wc.name as work_center_name,
 		       ws.name as shift_name, ws.expected_check_in, ws.expected_check_out,
 		       e.created_at as employee_created_at, e.is_active, p.hourly_rate
 		FROM users u
 		LEFT JOIN employees e ON e.user_id = u.id
 		LEFT JOIN positions p ON e.position_id = p.id
 		LEFT JOIN work_centers wc ON e.work_center_id = wc.id
 		LEFT JOIN work_shifts ws ON e.work_shift_id = ws.id
 		WHERE u.id = $1
 	`
	err := h.DB.Get(&profile, query, userID)
	if err != nil {
		// Fallback: If query fails (e.g. mapping error), try to at least get basic user info
		var basicUser struct {
			ID       int     `db:"id" json:"id"`
			Name     string  `db:"name" json:"name"`
			Email    string  `db:"email" json:"email"`
			Phone    *string `db:"phone" json:"phone"`
			PhotoURL *string `db:"photo_url" json:"photo_url"`
		}
		errBasic := h.DB.Get(&basicUser, "SELECT id, name, email, phone, photo_url FROM users WHERE id = $1", userID)
		if errBasic != nil {
			utils.GetLogger().Error("Error fetching basic profile", 
				zap.Int("userID", userID), 
				zap.Error(errBasic))
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
		}

		profile.ID = basicUser.ID
		profile.Name = basicUser.Name
		profile.Email = basicUser.Email
		profile.Phone = basicUser.Phone
		profile.PhotoURL = basicUser.PhotoURL
		profile.IsEmployee = false
	} else {
		profile.IsEmployee = profile.EmployeeCreatedAt != nil
	}

	if profile.PhotoURL != nil {
		fullURL := h.Storage.GetPublicURL(*profile.PhotoURL)
		profile.PhotoURL = &fullURL
	}

	return c.JSON(profile)
}

func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	var req struct {
		Name  string  `json:"name"`
		Email string  `json:"email"`
		Phone *string `json:"phone"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name and Email are required"})
	}

	_, err := h.DB.Exec("UPDATE users SET name = $1, email = $2, phone = $3, updated_at = NOW() WHERE id = $4", req.Name, req.Email, req.Phone, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update profile. Email might be in use."})
	}

	return c.JSON(fiber.Map{"message": "Profile updated successfully"})
}




