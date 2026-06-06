package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"attendance-api/internal/models"
	"attendance-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func (h *EmployeeAdminHandler) ListManagers(c *fiber.Ctx) error {
	var entities []models.User
	if err := h.DB.Select(&entities, "SELECT id, name, email, role_id FROM users WHERE role_id = (SELECT id FROM roles WHERE slug = $1)", models.RoleSlugManager); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	dtos := make([]models.UserDTO, 0)
	for _, u := range entities {
		dtos = append(dtos, models.MapUserToDTO(u, h.Cfg.R2PublicURL))
	}
	return c.JSON(dtos)
}

func (h *EmployeeAdminHandler) ListUnassignedUsers(c *fiber.Ctx) error {
	var entities []models.User
	query := `
		SELECT id, name, email 
		FROM users u 
		WHERE NOT EXISTS (SELECT 1 FROM employees e WHERE e.user_id = u.id)
	`
	if err := h.DB.Select(&entities, query); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	dtos := make([]models.UserDTO, 0)
	for _, u := range entities {
		dtos = append(dtos, models.MapUserToDTO(u, h.Cfg.R2PublicURL))
	}
	return c.JSON(dtos)
}

// EMPLOYEES
func (h *EmployeeAdminHandler) ListEmployees(c *fiber.Ctx) error {
	type EmployeeRich struct {
		models.Employee
		UserName     string  `db:"user_name"`
		Email        string  `db:"email"`
		Phone        *string `db:"phone"`
		CenterName   string  `db:"center_name"`
		ShiftName    *string `db:"shift_name"`
		PositionName string  `db:"position_name"`
		HourlyRate   float64 `db:"hourly_rate"`
		PhotoURL     *string `db:"photo_url"`
	}

	search := c.Query("search")
	centerID := c.Query("center_id")
	shiftID := c.Query("shift_id")
	positionID := c.Query("position_id")
	shiftType := c.Query("shift_type")

	query := `
		SELECT e.*, u.name as user_name, u.email, u.phone, u.photo_url, wc.name as center_name, ws.name as shift_name, p.name as position_name, p.hourly_rate
		FROM employees e
		JOIN users u ON e.user_id = u.id
		JOIN work_centers wc ON e.work_center_id = wc.id
		JOIN positions p ON e.position_id = p.id
		LEFT JOIN work_shifts ws ON e.work_shift_id = ws.id
		WHERE 1=1
	`
	args := make(map[string]interface{})

	if search != "" {
		query += " AND (u.name ILIKE :search OR u.email ILIKE :search OR u.phone ILIKE :search)"
		args["search"] = "%" + search + "%"
	}

	if centerID != "" && centerID != "all" {
		query += " AND e.work_center_id = :center_id"
		args["center_id"] = centerID
	}

	if shiftID != "" && shiftID != "all" {
		query += " AND e.work_shift_id = :shift_id"
		args["shift_id"] = shiftID
	}

	if positionID != "" && positionID != "all" {
		query += " AND e.position_id = :position_id"
		args["position_id"] = positionID
	}

	if shiftType != "" && shiftType != "all" {
		query += " AND ws.shift_type = :shift_type"
		args["shift_type"] = shiftType
	}

	query += " ORDER BY e.created_at DESC"

	var entities []EmployeeRich
	nQuery, nArgs, err := h.DB.BindNamed(query, args)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Query preparation failed", err)
	}

	if err := h.DB.Select(&entities, h.DB.Rebind(nQuery), nArgs...); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	dtos := make([]models.EmployeeDetailDTO, 0)
	for _, e := range entities {
		dtos = append(dtos, models.EmployeeDetailDTO{
			EmployeeDTO:  models.MapEmployeeToDTO(e.Employee),
			UserName:     e.UserName,
			Email:        e.Email,
			Phone:        e.Phone,
			JoinedAt:     e.CreatedAt.Format("Jan 2, 2006"),
			CenterName:   e.CenterName,
			ShiftName:    e.ShiftName,
			PositionName: e.PositionName,
			HourlyRate:   e.HourlyRate,
			PhotoURL: func() *string {
				if e.PhotoURL == nil { return nil }
				url := fmt.Sprintf("%s/%s", strings.TrimSuffix(h.Cfg.R2PublicURL, "/"), strings.TrimPrefix(*e.PhotoURL, "/"))
				return &url
			}(),
		})
	}

	return c.JSON(dtos)
}

func (h *EmployeeAdminHandler) CreateEmployee(c *fiber.Ctx) error {
	var emp models.Employee
	if err := c.BodyParser(&emp); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	now := time.Now()
	emp.CreatedAt = &now
	emp.UpdatedAt = &now
	emp.IsActive = true

	query := `INSERT INTO employees (user_id, work_center_id, work_shift_id, position_id, is_active, created_at, updated_at) 
			  VALUES (:user_id, :work_center_id, :work_shift_id, :position_id, :is_active, :created_at, :updated_at) RETURNING id`
	
	rows, err := tx.NamedQuery(query, emp)
	if err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}
	if rows.Next() {
		rows.Scan(&emp.ID)
	}
	rows.Close()

	// PROMOTE USER TO EMPLOYEE (Role ID 3) ONLY IF THEY ARE REGULAR USERS (Role ID 5)
	_, err = tx.Exec("UPDATE users SET role_id = (SELECT id FROM roles WHERE slug = $1) WHERE id = $2 AND role_id = (SELECT id FROM roles WHERE slug = $3)", models.RoleSlugEmployee, emp.UserID, models.RoleSlugUser)
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error promoting user role"})
	}

	userID := c.Locals("user_id").(int)
	if err := h.AuditService.LogAction(c.Context(), tx, userID, models.AuditActionCreateEmployee, "employee", emp.ID, nil, emp, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	return c.Status(fiber.StatusCreated).JSON(models.MapEmployeeToDTO(emp))
}

func (h *EmployeeAdminHandler) UpdateEmployee(c *fiber.Ctx) error {
	id := c.Params("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{Code: models.ErrInvalidID})
	}

	var emp models.Employee
	if err := c.BodyParser(&emp); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	userID := c.Locals("user_id").(int)

	var old models.Employee
	if err := tx.Get(&old, "SELECT * FROM employees WHERE id = $1", idInt); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Employee not found"})
	}

	_, err = tx.Exec("UPDATE employees SET work_center_id = $1, work_shift_id = $2, position_id = $3, is_active = $4, updated_at = $5 WHERE id = $6",
		emp.WorkCenterID, emp.WorkShiftID, emp.PositionID, emp.IsActive, time.Now(), idInt)
	
	if err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	if err := h.AuditService.LogAction(c.Context(), tx, userID, models.AuditActionUpdateEmployee, "employee", idInt, old, emp, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	return c.JSON(fiber.Map{"message": "Employee updated successfully"})
}

func (h *EmployeeAdminHandler) DeleteEmployee(c *fiber.Ctx) error {
	id := c.Params("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{Code: models.ErrInvalidID})
	}

	// Get user_id before deleting
	var userID int
	err = h.DB.Get(&userID, "SELECT user_id FROM employees WHERE id = $1", idInt)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Employee not found"})
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	// Revert role to User (Role ID 5) ONLY IF THEY ARE CURRENTLY EMPLOYEE (Role ID 3)
	_, err = tx.Exec("UPDATE users SET role_id = (SELECT id FROM roles WHERE slug = $1) WHERE id = $2 AND role_id = (SELECT id FROM roles WHERE slug = $3)", models.RoleSlugUser, userID, models.RoleSlugEmployee)
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error reverting user role"})
	}

	// Get OLD state for audit
	var old models.Employee
	tx.Get(&old, "SELECT * FROM employees WHERE id = $1", idInt)

	if _, err := tx.Exec("DELETE FROM employees WHERE id = $1", idInt); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot delete employee: history exists"})
	}

	// Log Action (Inside TX)
	adminID := c.Locals("user_id").(int)
	if err := h.AuditService.LogAction(c.Context(), tx, adminID, models.AuditActionDeleteEmployee, "employee", idInt, old, nil, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// POSITIONS

func (h *EmployeeAdminHandler) GetEmployeeDetails(c *fiber.Ctx) error {
	id := c.Params("id")

	// 1. Get Employee Rich Data
	var emp struct {
		models.Employee
		UserName     string  `db:"user_name"`
		Email        string  `db:"email"`
		Phone        *string `db:"phone"`
		CenterName   string  `db:"center_name"`
		ShiftName    *string `db:"shift_name"`
		PositionName string  `db:"position_name"`
		HourlyRate   float64 `db:"hourly_rate"`
	}
	empQuery := `
		SELECT e.*, u.name as user_name, u.email, u.phone, wc.name as center_name, ws.name as shift_name, p.name as position_name, p.hourly_rate
		FROM employees e
		JOIN users u ON e.user_id = u.id
		JOIN work_centers wc ON e.work_center_id = wc.id
		JOIN positions p ON e.position_id = p.id
		LEFT JOIN work_shifts ws ON e.work_shift_id = ws.id
		WHERE e.id = $1
	`
	if err := h.DB.Get(&emp, empQuery, id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Employee not found"})
	}

	// 2. Get Stats
	var stats models.EmployeeStats
	statsQuery := `
		SELECT 
			COUNT(id) as total_attendances,
			COALESCE(SUM(net_hours_worked), 0) as total_hours,
			COALESCE(SUM(daily_earnings), 0) as total_earnings
		FROM attendances
		WHERE employee_id = $1
	`
	h.DB.Get(&stats, statsQuery, id)

	// Count incidents
	incQuery := `
		SELECT COUNT(i.id) 
		FROM incidents i
		JOIN attendances a ON i.attendance_id = a.id
		WHERE a.employee_id = $1
	`
	h.DB.Get(&stats.TotalIncidents, incQuery, id)

	// 3. Get Recent Attendance
	type AttendanceRich struct {
		models.Attendance
		EmployeeName   string `db:"employee_name"`
		WorkCenterName string `db:"center_name"`
		IsLate         bool   `db:"is_late"`
	}
	attQuery := `
		SELECT 
			a.*, u.name as employee_name, wc.name as center_name,
			EXISTS(SELECT 1 FROM incidents i WHERE i.attendance_id = a.id AND i.type = $2) as is_late
		FROM attendances a
		JOIN employees e ON a.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		JOIN work_centers wc ON a.work_center_id = wc.id
		WHERE a.employee_id = $1
		ORDER BY a.check_in DESC
		LIMIT 50
	`
	var attendances []AttendanceRich
	h.DB.Select(&attendances, attQuery, id)

	attDTOs := make([]models.AttendanceDetailDTO, 0)
	for _, a := range attendances {
		attDTOs = append(attDTOs, models.MapAttendanceToDetailDTO(a.Attendance, a.EmployeeName, a.WorkCenterName, "", a.IsLate, h.Cfg.R2PublicURL))
	}

	return c.JSON(models.EmployeeFullDetailDTO{
		Employee: models.EmployeeDetailDTO{
			EmployeeDTO:  models.MapEmployeeToDTO(emp.Employee),
			UserName:     emp.UserName,
			Email:        emp.Email,
			JoinedAt:     emp.CreatedAt.Format("Jan 2, 2006"),
			CenterName:   emp.CenterName,
			ShiftName:    emp.ShiftName,
			PositionName: emp.PositionName,
			HourlyRate:   emp.HourlyRate,
		},
		Stats:            stats,
		RecentAttendance: attDTOs,
	})
}



