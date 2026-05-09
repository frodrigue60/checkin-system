package handlers

import (
	"strconv"
	"time"

	"attendance-api/internal/models"
	"attendance-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func (h *PositionHandler) ListPositions(c *fiber.Ctx) error {
	if data, found := h.Cache.Get("positions"); found {
		return c.JSON(data)
	}

	var entities []models.Position
	query := `
		SELECT p.*, COUNT(e.id) as employees_count 
		FROM positions p 
		LEFT JOIN employees e ON e.position_id = p.id 
		GROUP BY p.id 
		ORDER BY p.name
	`
	if err := h.DB.Select(&entities, query); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	dtos := make([]models.PositionDTO, 0)
	for _, p := range entities {
		dtos = append(dtos, models.MapPositionToDTO(p))
	}

	h.Cache.Set("positions", dtos, 5*time.Minute)
	return c.JSON(dtos)
}

func (h *PositionHandler) CreatePosition(c *fiber.Ctx) error {
	var pos models.Position
	if err := c.BodyParser(&pos); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	now := time.Now()
	pos.CreatedAt = &now
	pos.UpdatedAt = &now

	query := `INSERT INTO positions (name, hourly_rate, late_penalty_fee, out_of_range_fee, lunch_overstay_fee, created_at, updated_at) 
			  VALUES (:name, :hourly_rate, :late_penalty_fee, :out_of_range_fee, :lunch_overstay_fee, :created_at, :updated_at) RETURNING id`
	
	rows, err := tx.NamedQuery(query, pos)
	if err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}
	if rows.Next() {
		rows.Scan(&pos.ID)
	}
	rows.Close()

	userID := c.Locals("user_id").(int)
	if err := h.AuditService.LogAction(c.Context(), tx, userID, models.AuditActionCreatePosition, "position", pos.ID, nil, pos, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("positions")
	return c.Status(fiber.StatusCreated).JSON(models.MapPositionToDTO(pos))
}

func (h *PositionHandler) UpdatePosition(c *fiber.Ctx) error {
	id := c.Params("id")
	var pos models.Position
	if err := c.BodyParser(&pos); err != nil {
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

	var old models.Position
	if err := tx.Get(&old, "SELECT * FROM positions WHERE id = $1", idInt); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Position not found"})
	}

	_, err = tx.Exec("UPDATE positions SET name = $1, hourly_rate = $2, late_penalty_fee = $3, out_of_range_fee = $4, lunch_overstay_fee = $5, updated_at = $6 WHERE id = $7",
		pos.Name, pos.HourlyRate, pos.LatePenaltyFee, pos.OutOfRangeFee, pos.LunchOverstayFee, time.Now(), idInt)
	
	if err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	if err := h.AuditService.LogAction(c.Context(), tx, userID, models.AuditActionUpdatePosition, "position", idInt, old, pos, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("positions")
	return c.JSON(fiber.Map{"message": "Position updated successfully"})
}

func (h *PositionHandler) DeletePosition(c *fiber.Ctx) error {
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

	var old models.Position
	if err := tx.Get(&old, "SELECT * FROM positions WHERE id = $1", idInt); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Position not found"})
	}

	if _, err := tx.Exec("DELETE FROM positions WHERE id = $1", idInt); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot delete position"})
	}

	if err := h.AuditService.LogAction(c.Context(), tx, userID, models.AuditActionDeletePosition, "position", idInt, old, nil, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("positions")
	return c.SendStatus(fiber.StatusNoContent)
}

// HOLIDAYS

func (h *PositionHandler) GetPositionDetails(c *fiber.Ctx) error {
	id := c.Params("id")

	// 1. Get Position
	var pos models.Position
	if err := h.DB.Get(&pos, "SELECT * FROM positions WHERE id = $1", id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Position not found"})
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
		WHERE e.position_id = $1
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
		WHERE e.position_id = $1
		ORDER BY a.check_in DESC
		LIMIT 50
	`
	var attendances []AttendanceRich
	h.DB.Select(&attendances, attQuery, id)

	attDTOs := make([]models.AttendanceDetailDTO, 0)
	for _, a := range attendances {
		attDTOs = append(attDTOs, models.MapAttendanceToDetailDTO(a.Attendance, a.EmployeeName, a.WorkCenterName, "", false))
	}

	return c.JSON(models.PositionDetailDTO{
		Position:         models.MapPositionToDTO(pos),
		Employees:        empDTOs,
		RecentAttendance: attDTOs,
	})
}


