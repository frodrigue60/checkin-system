package handlers

import (
	"strconv"
	"time"

	"attendance-api/internal/models"
	"attendance-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func (h *ShiftHandler) ListShifts(c *fiber.Ctx) error {
	// 1. Check Cache
	if data, found := h.Cache.Get("shifts"); found {
		return c.JSON(data)
	}

	var entities []models.WorkShift
	if err := h.DB.Select(&entities, "SELECT * FROM work_shifts ORDER BY id"); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	dtos := make([]models.WorkShiftDTO, 0)
	for _, ws := range entities {
		dtos = append(dtos, models.MapWorkShiftToDTO(ws))
	}

	// 2. Save to Cache
	h.Cache.Set("shifts", dtos, 5*time.Minute)

	return c.JSON(dtos)
}

func (h *ShiftHandler) CreateShift(c *fiber.Ctx) error {
	var shift models.WorkShift
	if err := utils.ParseAndValidate(c, &shift); err != nil {
		return err
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	now := time.Now()
	shift.CreatedAt = &now
	shift.UpdatedAt = &now

	query := `INSERT INTO work_shifts (name, expected_check_in, expected_check_out, allowed_lunch_time, tolerance_time, is_night_shift, enforce_geofence, enforce_lateness, enforce_lunch_limit, shift_type, work_days, created_at, updated_at) 
			  VALUES (:name, :expected_check_in, :expected_check_out, :allowed_lunch_time, :tolerance_time, :is_night_shift, :enforce_geofence, :enforce_lateness, :enforce_lunch_limit, :shift_type, :work_days, :created_at, :updated_at) RETURNING id`
	
	rows, err := tx.NamedQuery(query, shift)
	if err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}
	if rows.Next() {
		rows.Scan(&shift.ID)
	}
	rows.Close()

	userID := c.Locals("user_id").(int)
	if err := h.AuditService.LogAction(c.Context(), tx, userID, models.AuditActionCreateShift, "work_shift", shift.ID, nil, shift, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("shifts")
	return c.Status(fiber.StatusCreated).JSON(models.MapWorkShiftToDTO(shift))
}

func (h *ShiftHandler) UpdateShift(c *fiber.Ctx) error {
	id := c.Params("id")
	var shift models.WorkShift
	if err := utils.ParseAndValidate(c, &shift); err != nil {
		return err
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

	var old models.WorkShift
	if err := tx.Get(&old, "SELECT * FROM work_shifts WHERE id = $1", idInt); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shift not found"})
	}

	_, err = tx.Exec("UPDATE work_shifts SET name = $1, expected_check_in = $2, expected_check_out = $3, allowed_lunch_time = $4, tolerance_time = $5, is_night_shift = $6, enforce_geofence = $7, enforce_lateness = $8, enforce_lunch_limit = $9, shift_type = $10, work_days = $11, updated_at = $12 WHERE id = $13",
		shift.Name, shift.ExpectedCheckIn, shift.ExpectedCheckOut, shift.AllowedLunchTime, shift.ToleranceTime, shift.IsNightShift, shift.EnforceGeofence, shift.EnforceLateness, shift.EnforceLunchLimit, shift.ShiftType, shift.WorkDays, time.Now(), idInt)
	
	if err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	if err := h.AuditService.LogAction(c.Context(), tx, userID, models.AuditActionUpdateShift, "work_shift", idInt, old, shift, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("shifts")
	return c.JSON(fiber.Map{"message": "Shift updated successfully"})
}

func (h *ShiftHandler) DeleteShift(c *fiber.Ctx) error {
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

	var old models.WorkShift
	if err := tx.Get(&old, "SELECT * FROM work_shifts WHERE id = $1", idInt); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shift not found"})
	}

	if _, err := tx.Exec("DELETE FROM work_shifts WHERE id = $1", id); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot delete shift: it might have linked employees"})
	}

	if err := h.AuditService.LogAction(c.Context(), tx, userID, models.AuditActionDeleteShift, "work_shift", idInt, old, nil, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("shifts")
	return c.SendStatus(fiber.StatusNoContent)
}

// USERS / MANAGERS

func (h *ShiftHandler) GetShiftDetails(c *fiber.Ctx) error {
	id := c.Params("id")

	// 1. Get Shift
	var shift models.WorkShift
	if err := h.DB.Get(&shift, "SELECT * FROM work_shifts WHERE id = $1", id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shift not found"})
	}

	// 2. Get Employees (Rich)
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
		WHERE e.work_shift_id = $1
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

	// 3. Get Recent Attendance (Rich)
	type AttendanceRich struct {
		models.Attendance
		EmployeeName  string `db:"employee_name"`
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
		WHERE a.work_shift_id = $1
		ORDER BY a.check_in DESC
		LIMIT 50
	`
	var attendances []AttendanceRich
	h.DB.Select(&attendances, attQuery, id)

	attDTOs := make([]models.AttendanceDetailDTO, 0)
	for _, a := range attendances {
		attDTOs = append(attDTOs, models.MapAttendanceToDetailDTO(a.Attendance, a.EmployeeName, a.WorkCenterName, "", false))
	}

	return c.JSON(models.WorkShiftDetailDTO{
		Shift:            models.MapWorkShiftToDTO(shift),
		Employees:        empDTOs,
		RecentAttendance: attDTOs,
	})
}


