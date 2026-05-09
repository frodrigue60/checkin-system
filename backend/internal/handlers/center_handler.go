package handlers

import (
	"attendance-api/internal/models"
	"attendance-api/internal/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// CENTERS — SOLID (S): This handler only manages WorkCenter CRUD operations.

func (h *CenterHandler) ListCenters(c *fiber.Ctx) error {
	if data, found := h.Cache.Get("centers"); found {
		return c.JSON(data)
	}

	var entities []models.WorkCenter
	if err := h.DB.Select(&entities, "SELECT * FROM work_centers ORDER BY created_at DESC"); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to fetch centers", err)
	}

	dtos := make([]models.WorkCenterDTO, 0)
	for _, wc := range entities {
		dtos = append(dtos, models.MapWorkCenterToDTO(wc))
	}

	h.Cache.Set("centers", dtos, 5*time.Minute)
	return c.JSON(dtos)
}

func (h *CenterHandler) CreateCenter(c *fiber.Ctx) error {
	var center models.WorkCenter
	if err := c.BodyParser(&center); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	now := time.Now()
	center.CreatedAt = &now
	center.UpdatedAt = &now

	query := `INSERT INTO work_centers (name, address, latitude, longitude, tolerance_radius_meters, manager_id, created_at, updated_at) 
			  VALUES (:name, :address, :latitude, :longitude, :tolerance_radius_meters, :manager_id, :created_at, :updated_at) RETURNING id`

	rows, err := tx.NamedQuery(query, center)
	if err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}
	if rows.Next() {
		rows.Scan(&center.ID)
	}
	rows.Close()

	userID := c.Locals("user_id").(int)
	if err := h.AuditService.LogAction(c.Context(), tx, userID, models.AuditActionCreateWorkCenter, "work_center", center.ID, nil, center, c.IP()); err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Error logging action", err)
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("centers")
	return c.Status(fiber.StatusCreated).JSON(models.MapWorkCenterToDTO(center))
}

func (h *CenterHandler) UpdateCenter(c *fiber.Ctx) error {
	id := c.Params("id")
	var center models.WorkCenter
	if err := c.BodyParser(&center); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	userID := c.Locals("user_id").(int)
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{Code: models.ErrInvalidID})
	}

	var old models.WorkCenter
	if err := tx.Get(&old, "SELECT * FROM work_centers WHERE id = $1", idInt); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Center not found"})
	}

	_, err = tx.Exec("UPDATE work_centers SET name = $1, address = $2, latitude = $3, longitude = $4, tolerance_radius_meters = $5, manager_id = $6, updated_at = $7 WHERE id = $8",
		center.Name, center.Address, center.Latitude, center.Longitude, center.ToleranceRadiusMeters, center.ManagerID, time.Now(), id)

	if err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	if err := h.AuditService.LogAction(c.Context(), tx, userID, models.AuditActionUpdateWorkCenter, "work_center", idInt, old, center, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("centers")
	return c.JSON(fiber.Map{"message": "Center updated successfully"})
}

func (h *CenterHandler) DeleteCenter(c *fiber.Ctx) error {
	id := c.Params("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{Code: models.ErrInvalidID})
	}
	userID := c.Locals("user_id").(int)

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	var old models.WorkCenter
	if err := tx.Get(&old, "SELECT * FROM work_centers WHERE id = $1", idInt); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Center not found"})
	}

	if _, err := tx.Exec("DELETE FROM work_centers WHERE id = $1", id); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot delete center: it might have linked employees or attendances"})
	}

	if err := h.AuditService.LogAction(c.Context(), tx, userID, models.AuditActionDeleteWorkCenter, "work_center", idInt, old, nil, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("centers")
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *CenterHandler) GetCenterDetails(c *fiber.Ctx) error {
	id := c.Params("id")

	var center models.WorkCenter
	if err := h.DB.Get(&center, "SELECT * FROM work_centers WHERE id = $1", id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Center not found"})
	}

	var managerPtr *models.UserDTO
	if center.ManagerID != nil {
		var manager models.User
		if err := h.DB.Get(&manager, "SELECT id, name, email, role_id, created_at FROM users WHERE id = $1", *center.ManagerID); err == nil {
			dto := models.MapUserToDTO(manager)
			managerPtr = &dto
		}
	}

	var employees []struct {
		models.Employee
		UserName     string  `db:"user_name"`
		CenterName   string  `db:"center_name"`
		ShiftName    *string `db:"shift_name"`
		PositionName string  `db:"position_name"`
		HourlyRate   float64 `db:"hourly_rate"`
	}
	empQuery := `
		SELECT e.*, u.name as user_name, wc.name as center_name, ws.name as shift_name, p.name as position_name, p.hourly_rate
		FROM employees e
		JOIN users u ON e.user_id = u.id
		JOIN work_centers wc ON e.work_center_id = wc.id
		JOIN positions p ON e.position_id = p.id
		LEFT JOIN work_shifts ws ON e.work_shift_id = ws.id
		WHERE e.work_center_id = $1
		ORDER BY u.name ASC
	`
	h.DB.Select(&employees, empQuery, id)

	empDTOs := make([]models.EmployeeDetailDTO, 0)
	for _, e := range employees {
		empDTOs = append(empDTOs, models.EmployeeDetailDTO{
			EmployeeDTO:  models.MapEmployeeToDTO(e.Employee),
			UserName:     e.UserName,
			CenterName:   e.CenterName,
			ShiftName:    e.ShiftName,
			PositionName: e.PositionName,
			HourlyRate:   e.HourlyRate,
		})
	}

	type AttendanceRich struct {
		models.Attendance
		EmployeeName   string `db:"employee_name"`
		WorkCenterName string `db:"center_name"`
	}
	attQuery := `
		SELECT 
			a.id, a.employee_id, a.work_shift_id, a.work_center_id, 
			a.check_in, a.lunch_start, a.lunch_end, a.check_out,
			a.check_in_latitude, a.check_in_longitude, a.check_out_latitude, a.check_out_longitude,
			COALESCE(a.net_hours_worked, 0) as net_hours_worked, 
			COALESCE(a.daily_earnings, 0) as daily_earnings,
			a.created_at, a.updated_at,
			u.name as employee_name, wc.name as center_name
		FROM attendances a
		JOIN employees e ON a.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		JOIN work_centers wc ON a.work_center_id = wc.id
		WHERE a.work_center_id = $1
		ORDER BY a.check_in DESC
		LIMIT 50
	`
	var attendances []AttendanceRich
	h.DB.Select(&attendances, attQuery, id)

	attDTOs := make([]models.AttendanceDetailDTO, 0)
	for _, a := range attendances {
		attDTOs = append(attDTOs, models.MapAttendanceToDetailDTO(a.Attendance, a.EmployeeName, a.WorkCenterName, "", false))
	}

	return c.JSON(models.WorkCenterDetailDTO{
		Center:           models.MapWorkCenterToDTO(center),
		Manager:          managerPtr,
		Employees:        empDTOs,
		RecentAttendance: attDTOs,
	})
}
