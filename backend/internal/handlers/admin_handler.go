package handlers

import (
	"attendance-api/internal/models"
	"attendance-api/internal/services"
	"attendance-api/internal/utils"
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
)

type AttendanceRich struct {
	models.Attendance
	EmployeeName   string `db:"employee_name"`
	WorkCenterName string `db:"center_name"`
	PositionName   string `db:"position_name"`
	IsLate         bool   `db:"is_late"`
}

type AdminHandler struct {
	DB                *sqlx.DB
	PDFService        *services.PDFService
	AttendanceService *services.AttendanceService
	AuditService      *services.AuditService
	ReportService     *services.ReportService
	AlertService      *services.AlertService
	JustificationService *services.JustificationService
	Cache                *cache.Cache
}

// CENTERS
func (h *AdminHandler) ListCenters(c *fiber.Ctx) error {
	// 1. Check Cache
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

	// 2. Save to Cache
	h.Cache.Set("centers", dtos, 5*time.Minute)

	return c.JSON(dtos)
}

func (h *AdminHandler) CreateCenter(c *fiber.Ctx) error {
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
	if err := h.AuditService.LogActionTx(tx, userID, models.AuditActionCreateWorkCenter, "work_center", center.ID, nil, center, c.IP()); err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Error logging action", err)
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("centers")
	return c.Status(fiber.StatusCreated).JSON(models.MapWorkCenterToDTO(center))
}

func (h *AdminHandler) UpdateCenter(c *fiber.Ctx) error {
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
	idInt, _ := strconv.Atoi(id)

	// Get OLD state
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

	if err := h.AuditService.LogActionTx(tx, userID, models.AuditActionUpdateWorkCenter, "work_center", idInt, old, center, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("centers")
	return c.JSON(fiber.Map{"message": "Center updated successfully"})
}

func (h *AdminHandler) DeleteCenter(c *fiber.Ctx) error {
	id := c.Params("id")
	idInt, _ := strconv.Atoi(id)
	userID := c.Locals("user_id").(int)

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	// Get OLD state
	var old models.WorkCenter
	if err := tx.Get(&old, "SELECT * FROM work_centers WHERE id = $1", idInt); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Center not found"})
	}

	if _, err := tx.Exec("DELETE FROM work_centers WHERE id = $1", id); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Cannot delete center: it might have linked employees or attendances"})
	}

	if err := h.AuditService.LogActionTx(tx, userID, models.AuditActionDeleteWorkCenter, "work_center", idInt, old, nil, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("centers")
	return c.SendStatus(fiber.StatusNoContent)
}

// SHIFTS
func (h *AdminHandler) ListShifts(c *fiber.Ctx) error {
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

func (h *AdminHandler) CreateShift(c *fiber.Ctx) error {
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
	if err := h.AuditService.LogActionTx(tx, userID, models.AuditActionCreateShift, "work_shift", shift.ID, nil, shift, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("shifts")
	return c.Status(fiber.StatusCreated).JSON(models.MapWorkShiftToDTO(shift))
}

func (h *AdminHandler) UpdateShift(c *fiber.Ctx) error {
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
	idInt, _ := strconv.Atoi(id)

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

	if err := h.AuditService.LogActionTx(tx, userID, models.AuditActionUpdateShift, "work_shift", idInt, old, shift, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("shifts")
	return c.JSON(fiber.Map{"message": "Shift updated successfully"})
}

func (h *AdminHandler) DeleteShift(c *fiber.Ctx) error {
	id := c.Params("id")
	idInt, _ := strconv.Atoi(id)
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

	if err := h.AuditService.LogActionTx(tx, userID, models.AuditActionDeleteShift, "work_shift", idInt, old, nil, c.IP()); err != nil {
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
func (h *AdminHandler) ListManagers(c *fiber.Ctx) error {
	var entities []models.User
	if err := h.DB.Select(&entities, "SELECT id, name, email, role_id FROM users WHERE role_id = 2"); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	dtos := make([]models.UserDTO, 0)
	for _, u := range entities {
		dtos = append(dtos, models.MapUserToDTO(u))
	}
	return c.JSON(dtos)
}

func (h *AdminHandler) ListUnassignedUsers(c *fiber.Ctx) error {
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
		dtos = append(dtos, models.MapUserToDTO(u))
	}
	return c.JSON(dtos)
}

// EMPLOYEES
func (h *AdminHandler) ListEmployees(c *fiber.Ctx) error {
	type EmployeeRich struct {
		models.Employee
		UserName     string  `db:"user_name"`
		Email        string  `db:"email"`
		Phone        *string `db:"phone"`
		CenterName   string  `db:"center_name"`
		ShiftName    *string `db:"shift_name"`
		PositionName string  `db:"position_name"`
		HourlyRate   float64 `db:"hourly_rate"`
	}

	search := c.Query("search")
	centerID := c.Query("center_id")
	shiftID := c.Query("shift_id")
	positionID := c.Query("position_id")
	shiftType := c.Query("shift_type")

	query := `
		SELECT e.*, u.name as user_name, u.email, u.phone, wc.name as center_name, ws.name as shift_name, p.name as position_name, p.hourly_rate
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
		})
	}

	return c.JSON(dtos)
}

func (h *AdminHandler) CreateEmployee(c *fiber.Ctx) error {
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
	_, err = tx.Exec("UPDATE users SET role_id = 3 WHERE id = $1 AND role_id = 5", emp.UserID)
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error promoting user role"})
	}

	userID := c.Locals("user_id").(int)
	if err := h.AuditService.LogActionTx(tx, userID, models.AuditActionCreateEmployee, "employee", emp.ID, nil, emp, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	return c.Status(fiber.StatusCreated).JSON(models.MapEmployeeToDTO(emp))
}

func (h *AdminHandler) UpdateEmployee(c *fiber.Ctx) error {
	id := c.Params("id")
	var emp models.Employee
	if err := c.BodyParser(&emp); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	userID := c.Locals("user_id").(int)
	idInt, _ := strconv.Atoi(id)

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

	if err := h.AuditService.LogActionTx(tx, userID, models.AuditActionUpdateEmployee, "employee", idInt, old, emp, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	return c.JSON(fiber.Map{"message": "Employee updated successfully"})
}

func (h *AdminHandler) DeleteEmployee(c *fiber.Ctx) error {
	id := c.Params("id")
	idInt, _ := strconv.Atoi(id)

	// Get user_id before deleting
	var userID int
	err := h.DB.Get(&userID, "SELECT user_id FROM employees WHERE id = $1", idInt)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Employee not found"})
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	// Revert role to User (Role ID 5) ONLY IF THEY ARE CURRENTLY EMPLOYEE (Role ID 3)
	_, err = tx.Exec("UPDATE users SET role_id = 5 WHERE id = $1 AND role_id = 3", userID)
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
	if err := h.AuditService.LogActionTx(tx, adminID, models.AuditActionDeleteEmployee, "employee", idInt, old, nil, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// POSITIONS
func (h *AdminHandler) ListPositions(c *fiber.Ctx) error {
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

func (h *AdminHandler) CreatePosition(c *fiber.Ctx) error {
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
	if err := h.AuditService.LogActionTx(tx, userID, models.AuditActionCreatePosition, "position", pos.ID, nil, pos, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("positions")
	return c.Status(fiber.StatusCreated).JSON(models.MapPositionToDTO(pos))
}

func (h *AdminHandler) UpdatePosition(c *fiber.Ctx) error {
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
	idInt, _ := strconv.Atoi(id)

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

	if err := h.AuditService.LogActionTx(tx, userID, models.AuditActionUpdatePosition, "position", idInt, old, pos, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("positions")
	return c.JSON(fiber.Map{"message": "Position updated successfully"})
}

func (h *AdminHandler) DeletePosition(c *fiber.Ctx) error {
	id := c.Params("id")
	idInt, _ := strconv.Atoi(id)
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

	if err := h.AuditService.LogActionTx(tx, userID, models.AuditActionDeletePosition, "position", idInt, old, nil, c.IP()); err != nil {
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
func (h *AdminHandler) ListHolidays(c *fiber.Ctx) error {
	var entities []models.Holiday
	if err := h.DB.Select(&entities, "SELECT * FROM holidays ORDER BY date"); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	dtos := make([]models.HolidayDTO, 0)
	for _, hol := range entities {
		dtos = append(dtos, models.MapHolidayToDTO(hol))
	}
	return c.JSON(dtos)
}

func (h *AdminHandler) CreateHoliday(c *fiber.Ctx) error {
	var holiday models.Holiday
	if err := c.BodyParser(&holiday); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	now := time.Now()
	holiday.CreatedAt = &now
	holiday.UpdatedAt = &now

	query := `INSERT INTO holidays (name, date, description, type, created_at, updated_at) 
			  VALUES (:name, :date, :description, :type, :created_at, :updated_at) RETURNING id`
	
	rows, err := tx.NamedQuery(query, holiday)
	if err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}
	if rows.Next() {
		rows.Scan(&holiday.ID)
	}
	rows.Close()

	userID := c.Locals("user_id").(int)
	if err := h.AuditService.LogActionTx(tx, userID, models.AuditActionCreateHoliday, "holiday", holiday.ID, nil, holiday, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("holidays")
	return c.Status(fiber.StatusCreated).JSON(models.MapHolidayToDTO(holiday))
}

func (h *AdminHandler) UpdateHoliday(c *fiber.Ctx) error {
	id := c.Params("id")
	var holiday models.Holiday
	if err := c.BodyParser(&holiday); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	userID := c.Locals("user_id").(int)
	idInt, _ := strconv.Atoi(id)

	var old models.Holiday
	if err := tx.Get(&old, "SELECT * FROM holidays WHERE id = $1", idInt); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Holiday not found"})
	}

	_, err = tx.Exec("UPDATE holidays SET name = $1, date = $2, description = $3, type = $4, updated_at = $5 WHERE id = $6",
		holiday.Name, holiday.Date, holiday.Description, holiday.Type, time.Now(), idInt)
	
	if err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	if err := h.AuditService.LogActionTx(tx, userID, models.AuditActionUpdateHoliday, "holiday", idInt, old, holiday, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("holidays")
	return c.JSON(fiber.Map{"message": "Holiday updated successfully"})
}

func (h *AdminHandler) DeleteHoliday(c *fiber.Ctx) error {
	id := c.Params("id")
	idInt, _ := strconv.Atoi(id)
	userID := c.Locals("user_id").(int)

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	var old models.Holiday
	if err := tx.Get(&old, "SELECT * FROM holidays WHERE id = $1", idInt); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Holiday not found"})
	}

	if _, err := tx.Exec("DELETE FROM holidays WHERE id = $1", idInt); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error deleting holiday"})
	}

	if err := h.AuditService.LogActionTx(tx, userID, models.AuditActionDeleteHoliday, "holiday", idInt, old, nil, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("holidays")
	return c.SendStatus(fiber.StatusNoContent)
}

// Helper to build attendance filters
func (h *AdminHandler) buildAttendanceFilters(c *fiber.Ctx) ([]string, []interface{}) {
	start := c.Query("start")
	end := c.Query("end")
	status := c.Query("status", "all")

	whereClauses := []string{}
	params := []interface{}{}

	// Security: Manager/Supervisor Filtering
	// If not admin, restrict to work centers where the user is the manager
	roleSlug, _ := c.Locals("role_slug").(string)
	userID, _ := c.Locals("user_id").(int)
	if roleSlug != "" && roleSlug != "admin" {
		whereClauses = append(whereClauses, fmt.Sprintf("a.work_center_id IN (SELECT id FROM work_centers WHERE manager_id = $%d)", len(params)+1))
		params = append(params, userID)
	}

	if start != "" && end != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("a.check_in::date >= $%d AND a.check_in::date <= $%d", len(params)+1, len(params)+2))
		params = append(params, start, end)
	}

	if status == "present" {
		whereClauses = append(whereClauses, "a.check_in IS NOT NULL AND a.check_out IS NULL AND a.is_absence = false")
	} else if status == "late" {
		whereClauses = append(whereClauses, fmt.Sprintf("EXISTS(SELECT 1 FROM incidents i WHERE i.attendance_id = a.id AND i.type = '%s')", models.IncidentTypeLate))
	} else if status == "absence" {
		whereClauses = append(whereClauses, "a.is_absence = true")
	}

	// New Filters
	centerID := c.QueryInt("center_id", 0)
	if centerID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("a.work_center_id = $%d", len(params)+1))
		params = append(params, centerID)
	}

	positionID := c.QueryInt("position_id", 0)
	if positionID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("e.position_id = $%d", len(params)+1))
		params = append(params, positionID)
	}

	shiftID := c.QueryInt("shift_id", 0)
	if shiftID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("a.work_shift_id = $%d", len(params)+1))
		params = append(params, shiftID)
	}

	shiftType := c.Query("shift_type")
	if shiftType != "" && shiftType != "all" {
		whereClauses = append(whereClauses, fmt.Sprintf("ws.shift_type = $%d", len(params)+1))
		params = append(params, shiftType)
	}

	search := c.Query("search")
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("u.name ILIKE $%d", len(params)+1))
		params = append(params, "%"+search+"%")
	}

	return whereClauses, params
}

// ATTENDANCES (Admin Global History)
func (h *AdminHandler) ListAttendances(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)
	if page < 1 { page = 1 }
	if limit < 1 { limit = 50 }
	offset := (page - 1) * limit

	whereClauses, params := h.buildAttendanceFilters(c)

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 1. Get total count for pagination
	var total int
	countQuery := "SELECT COUNT(*) FROM attendances a LEFT JOIN work_shifts ws ON a.work_shift_id = ws.id" + whereClause
	if len(params) > 0 {
		h.DB.Get(&total, countQuery, params...)
	} else {
		h.DB.Get(&total, countQuery)
	}

	// 2. Fetch paginated data
	incidentParamIdx := len(params) + 1
	limitParamIdx := len(params) + 2
	offsetParamIdx := len(params) + 3

	limitOffsetClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", limitParamIdx, offsetParamIdx)
	queryTemplate := `
		SELECT 
			a.id, a.employee_id, a.work_shift_id, a.work_center_id, 
			a.check_in, a.lunch_start, a.lunch_end, a.check_out,
			a.check_in_latitude, a.check_in_longitude, a.check_out_latitude, a.check_out_longitude,
			a.is_absence, a.absence_reason,
			COALESCE(a.net_hours_worked, 0) as net_hours_worked, 
			COALESCE(a.daily_earnings, 0) as daily_earnings,
			a.created_at, a.updated_at,
			COALESCE(u.name, 'Desconocido') as employee_name, 
			COALESCE(wc.name, 'Sin Sede') as center_name,
			COALESCE(p.name, 'Sin Puesto') as position_name,
			EXISTS(SELECT 1 FROM incidents i WHERE i.attendance_id = a.id AND i.type = $%d) as is_late
		FROM attendances a
		LEFT JOIN employees e ON a.employee_id = e.id
		LEFT JOIN users u ON e.user_id = u.id
		LEFT JOIN work_centers wc ON a.work_center_id = wc.id
		LEFT JOIN positions p ON e.position_id = p.id
		LEFT JOIN work_shifts ws ON a.work_shift_id = ws.id
		` + whereClause + `
		ORDER BY a.created_at DESC, a.id DESC
		` + limitOffsetClause
	
	query := fmt.Sprintf(queryTemplate, incidentParamIdx)
	queryParams := append(params, models.IncidentTypeLate, limit, offset)
	
	var entities []AttendanceRich
	err := h.DB.Select(&entities, query, queryParams...)

	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	dtos := make([]models.AttendanceDetailDTO, 0)
	for _, a := range entities {
		dtos = append(dtos, models.MapAttendanceToDetailDTO(a.Attendance, a.EmployeeName, a.WorkCenterName, a.PositionName, a.IsLate))
	}

	totalPages := (total + limit - 1) / limit

	return c.JSON(models.PaginatedResponse{
		Data:       dtos,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

func (h *AdminHandler) GetCenterDetails(c *fiber.Ctx) error {
	id := c.Params("id")

	// 1. Get Center
	var center models.WorkCenter
	if err := h.DB.Get(&center, "SELECT * FROM work_centers WHERE id = $1", id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Center not found"})
	}

	// 2. Get Manager
	var managerPtr *models.UserDTO
	if center.ManagerID != nil {
		var manager models.User
		if err := h.DB.Get(&manager, "SELECT id, name, email, role_id, created_at FROM users WHERE id = $1", *center.ManagerID); err == nil {
			dto := models.MapUserToDTO(manager)
			managerPtr = &dto
		}
	}

	// 3. Get Employees (Rich)
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

	// 4. Get Recent Attendance (Rich)
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

func (h *AdminHandler) GetPositionDetails(c *fiber.Ctx) error {
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

func (h *AdminHandler) GetShiftDetails(c *fiber.Ctx) error {
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

func (h *AdminHandler) GetEmployeeDetails(c *fiber.Ctx) error {
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
		attDTOs = append(attDTOs, models.MapAttendanceToDetailDTO(a.Attendance, a.EmployeeName, a.WorkCenterName, "", a.IsLate))
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


func (h *AdminHandler) DeleteAttendance(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := c.Locals("user_id").(int)
	idInt, _ := strconv.Atoi(id)

	// 0. Get OLD record for audit
	var oldAtt models.Attendance
	h.DB.Get(&oldAtt, "SELECT * FROM attendances WHERE id = $1", id)

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	// 1. Delete associated incidents
	if _, err := tx.Exec("DELETE FROM incidents WHERE attendance_id = $1", id); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error deleting related incidents"})
	}

	// 2. Delete attendance record
	if _, err := tx.Exec("DELETE FROM attendances WHERE id = $1", id); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error deleting attendance record"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit deletion"})
	}

	// Invalidate reports
	if oldAtt.CheckIn != nil {
		h.ReportService.InvalidateReports(oldAtt.EmployeeID, *oldAtt.CheckIn)
	}

	// 4. Log Action
	h.AuditService.LogAction(userID, models.AuditActionDeleteAttendance, "attendance", idInt, oldAtt, nil, c.IP())

	return c.SendStatus(fiber.StatusNoContent)
}

// DASHBOARD STATS
func (h *AdminHandler) GetDashboardStats(c *fiber.Ctx) error {
	var stats struct {
		TotalEmployees int `db:"total_employees" json:"total_employees"`
		TotalCenters   int `db:"total_centers" json:"total_centers"`
	}

	// 1. Get Basic Counts
	err := h.DB.Get(&stats, "SELECT (SELECT COUNT(*) FROM employees WHERE is_active = true) as total_employees, (SELECT COUNT(*) FROM work_centers) as total_centers")
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	// 2. Get Recent Incidents with Context
	type IncidentWithContext struct {
		ID           int       `db:"id" json:"id"`
		EmployeeName string    `db:"employee_name" json:"employee_name"`
		CenterName   string    `db:"center_name" json:"center_name"`
		Type         string    `db:"type" json:"type"`
		CreatedAt    time.Time `db:"created_at" json:"created_at"`
	}

	var incidents []IncidentWithContext
	incidentQuery := `
		SELECT i.id, u.name as employee_name, COALESCE(wc.name, 'Sede no identificada') as center_name, i.type, i.created_at
		FROM incidents i
		JOIN attendances a ON i.attendance_id = a.id
		JOIN employees e ON a.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		LEFT JOIN work_centers wc ON a.work_center_id = wc.id
		ORDER BY i.created_at DESC
		LIMIT 5
	`
	_ = h.DB.Select(&incidents, incidentQuery)

	// 3. System Alerts
	var alerts []models.SystemAlert
	h.DB.Select(&alerts, "SELECT * FROM system_alerts WHERE is_read = false ORDER BY created_at DESC LIMIT 10")

	// 4. Pending Justifications
	var justifications []models.Justification
	h.DB.Select(&justifications, "SELECT * FROM justifications WHERE status = 'pending' ORDER BY created_at DESC LIMIT 10")

	// 5. Compliance Trend (last 7 days)
	type TrendPoint struct {
		Date       string  `json:"date"`
		Attendance int     `json:"attendance"`
		Incidents  int     `json:"incidents"`
		Compliance float64 `json:"compliance"`
	}

	var trend []TrendPoint
	trendQuery := `
		SELECT 
			d::date::text as date,
			(SELECT COUNT(*) FROM attendances a WHERE a.check_in::date = d::date) as attendance,
			(SELECT COUNT(*) FROM incidents i WHERE i.created_at::date = d::date) as incidents
		FROM generate_series(CURRENT_DATE - INTERVAL '6 days', CURRENT_DATE, '1 day'::interval) d
		ORDER BY d ASC
	`
	_ = h.DB.Select(&trend, trendQuery)

	// Calculate compliance for each point
	var totalComp float64
	for i, p := range trend {
		if p.Attendance > 0 {
			trend[i].Compliance = 100.0 - (float64(p.Incidents) / float64(p.Attendance) * 100.0)
			if trend[i].Compliance < 0 { trend[i].Compliance = 0 }
		} else {
			trend[i].Compliance = 100.0
		}
		totalComp += trend[i].Compliance
	}

	complianceRate := totalComp / 7.0

	return c.JSON(fiber.Map{
		"total_employees":   stats.TotalEmployees,
		"total_centers":     stats.TotalCenters,
		"recent_incidents":  incidents,
		"compliance_rate":   fmt.Sprintf("%.1f", complianceRate),
		"compliance_trend":  trend,
		"alerts":            alerts,
		"justifications":    justifications,
	})
}

func (h *AdminHandler) ExportAttendancesCSV(c *fiber.Ctx) error {
	whereClauses, params := h.buildAttendanceFilters(c)
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query := `
		SELECT 
			a.check_in, a.check_out, a.is_absence, a.absence_reason,
			COALESCE(a.net_hours_worked, 0) as net_hours_worked, 
			COALESCE(a.daily_earnings, 0) as daily_earnings,
			COALESCE(u.name, 'N/A') as employee_name, 
			COALESCE(wc.name, 'N/A') as center_name,
			COALESCE(p.name, 'N/A') as position_name
		FROM attendances a
		LEFT JOIN employees e ON a.employee_id = e.id
		LEFT JOIN users u ON e.user_id = u.id
		LEFT JOIN work_centers wc ON a.work_center_id = wc.id
		LEFT JOIN positions p ON e.position_id = p.id
		` + whereClause + `
		ORDER BY a.created_at DESC
	`

	type ExportRow struct {
		CheckIn       *time.Time `db:"check_in"`
		CheckOut      *time.Time `db:"check_out"`
		IsAbsence     bool       `db:"is_absence"`
		AbsenceReason *string    `db:"absence_reason"`
		NetHours      float64    `db:"net_hours_worked"`
		Earnings      float64    `db:"daily_earnings"`
		EmployeeName  string     `db:"employee_name"`
		CenterName    string     `db:"center_name"`
		PositionName  string     `db:"position_name"`
	}

	var rows []ExportRow
	queryParams := append(params, models.IncidentTypeLate)
	if err := h.DB.Select(&rows, query, queryParams...); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	b := &bytes.Buffer{}
	// Add UTF-8 BOM for Excel
	b.Write([]byte{0xEF, 0xBB, 0xBF})
	
	w := csv.NewWriter(b)

	// Header
	w.Write([]string{"Empleado", "Sede", "Puesto", "Fecha/Entrada", "Salida", "Horas", "Ganancia", "Tipo", "Nota"})

	for _, r := range rows {
		dateStr := "---"
		if r.CheckIn != nil {
			dateStr = r.CheckIn.Format("2006-01-02 15:04:05")
		}
		checkoutStr := "---"
		if r.CheckOut != nil {
			checkoutStr = r.CheckOut.Format("15:04:05")
		}
		
		tipo := "Asistencia"
		if r.IsAbsence {
			tipo = "Ausencia"
			dateStr = "N/A"
		}

		note := ""
		if r.AbsenceReason != nil {
			note = *r.AbsenceReason
		}

		w.Write([]string{
			r.EmployeeName,
			r.CenterName,
			r.PositionName,
			dateStr,
			checkoutStr,
			fmt.Sprintf("%.2f", r.NetHours),
			fmt.Sprintf("%.2f", r.Earnings),
			tipo,
			note,
		})
	}

	w.Flush()

	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=asistencias_"+time.Now().Format("2006-01-02")+".csv")
	return c.Send(b.Bytes())
}

func (h *AdminHandler) ExportAttendancesPDF(c *fiber.Ctx) error {
	whereClauses, params := h.buildAttendanceFilters(c)
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query := `
		SELECT 
			a.id, a.check_in, a.check_out, a.is_absence, a.absence_reason,
			COALESCE(a.net_hours_worked, 0) as net_hours_worked, 
			COALESCE(a.daily_earnings, 0) as daily_earnings,
			COALESCE(u.name, 'N/A') as employee_name, 
			COALESCE(wc.name, 'N/A') as center_name,
			COALESCE(p.name, 'N/A') as position_name,
			EXISTS(SELECT 1 FROM incidents i WHERE i.attendance_id = a.id AND i.type = $3) as is_late
		FROM attendances a
		LEFT JOIN employees e ON a.employee_id = e.id
		LEFT JOIN users u ON e.user_id = u.id
		LEFT JOIN work_centers wc ON a.work_center_id = wc.id
		LEFT JOIN positions p ON e.position_id = p.id
		` + whereClause + `
		ORDER BY a.created_at DESC
		LIMIT 1000
	`

	type ExportRow struct {
		ID            int        `db:"id"`
		CheckIn       *time.Time `db:"check_in"`
		CheckOut      *time.Time `db:"check_out"`
		IsAbsence     bool       `db:"is_absence"`
		AbsenceReason *string    `db:"absence_reason"`
		NetHours      float64    `db:"net_hours_worked"`
		Earnings      float64    `db:"daily_earnings"`
		EmployeeName  string     `db:"employee_name"`
		CenterName    string     `db:"center_name"`
		PositionName  string     `db:"position_name"`
		IsLate        bool       `db:"is_late"`
	}

	var rows []ExportRow
	queryParams := append(params, models.IncidentTypeLate)
	if err := h.DB.Select(&rows, query, queryParams...); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	dtos := make([]models.AttendanceExportDTO, 0)
	for _, r := range rows {
		dateStr := "Ausencia"
		if !r.IsAbsence && r.CheckIn != nil {
			dateStr = r.CheckIn.Format("2006-01-02 15:04")
		}
		checkoutStr := "---"
		if r.CheckOut != nil {
			checkoutStr = r.CheckOut.Format("15:04")
		}
		
		reason := ""
		if r.AbsenceReason != nil {
			reason = *r.AbsenceReason
		}

		dtos = append(dtos, models.AttendanceExportDTO{
			ID:            r.ID,
			EmployeeName:  r.EmployeeName,
			CenterName:    r.CenterName,
			PositionName:  r.PositionName,
			CheckIn:       dateStr,
			CheckOut:      checkoutStr,
			Hours:         r.NetHours,
			Earnings:      r.Earnings,
			IsLate:        r.IsLate,
			IsAbsence:     r.IsAbsence,
			AbsenceReason: reason,
		})
	}

	pdf, err := h.PDFService.GenerateAttendanceLogPDF(dtos, "es")
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=historial_asistencias.pdf")
	return c.Send(pdf)
}

func (h *AdminHandler) GetAttendanceDetails(c *fiber.Ctx) error {
	id := c.Params("id")

	// 1. Get Attendance with basic names
	query := `
		SELECT 
			a.*,
			u.name as employee_name, 
			COALESCE(wc.name, 'Sin Sede') as center_name,
			COALESCE(p.name, 'Sin Puesto') as position_name,
			EXISTS(SELECT 1 FROM incidents i WHERE i.attendance_id = a.id AND i.type = $2) as is_late
		FROM attendances a
		LEFT JOIN employees e ON a.employee_id = e.id
		LEFT JOIN users u ON e.user_id = u.id
		LEFT JOIN work_centers wc ON a.work_center_id = wc.id
		LEFT JOIN positions p ON e.position_id = p.id
		WHERE a.id = $1
	`
	var attRich AttendanceRich
	if err := h.DB.Get(&attRich, query, id, models.IncidentTypeLate); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Asistencia no encontrada"})
	}

	// 2. Get WorkCenter
	var center models.WorkCenter
	if attRich.WorkCenterID != nil {
		h.DB.Get(&center, "SELECT * FROM work_centers WHERE id = $1", attRich.WorkCenterID)
	}

	// 3. Get Shift (if exists)
	var shiftDTO *models.WorkShiftDTO
	if attRich.WorkShiftID != nil {
		var shift models.WorkShift
		if err := h.DB.Get(&shift, "SELECT * FROM work_shifts WHERE id = $1", *attRich.WorkShiftID); err == nil {
			dto := models.MapWorkShiftToDTO(shift)
			shiftDTO = &dto
		}
	}

	// 4. Get Incidents
	var incidents []models.Incident
	err := h.DB.Unsafe().Select(&incidents, "SELECT * FROM incidents WHERE attendance_id = $1 ORDER BY created_at ASC", id)
	if err != nil {
		fmt.Printf("[DEBUG] Error fetching incidents for att %s: %v\n", id, err)
	}

	incidentDTOs := make([]models.IncidentDTO, 0)
	for _, i := range incidents {
		incidentDTOs = append(incidentDTOs, models.MapIncidentToDTO(i))
	}

	return c.JSON(models.AttendanceFullDetailDTO{
		Attendance: models.MapAttendanceToDetailDTO(attRich.Attendance, attRich.EmployeeName, attRich.WorkCenterName, attRich.PositionName, attRich.IsLate),
		WorkCenter: models.MapWorkCenterToDTO(center),
		Shift:      shiftDTO,
		Incidents:  incidentDTOs,
	})
}

// UpdateAttendance allows manual override of timestamps by admin
func (h *AdminHandler) UpdateAttendance(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	var req struct {
		CheckIn       *time.Time `json:"check_in"`
		LunchStart    *time.Time `json:"lunch_start"`
		LunchEnd      *time.Time `json:"lunch_end"`
		CheckOut      *time.Time `json:"check_out"`
		IsAbsence     bool       `json:"is_absence"`
		AbsenceReason *string    `json:"absence_reason"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID := c.Locals("user_id").(int)

	// 0. Get OLD state for audit
	var oldAtt models.Attendance
	h.DB.Get(&oldAtt, "SELECT * FROM attendances WHERE id = $1", id)

	// Update timestamps and absence status
	_, err = h.DB.Exec(`
		UPDATE attendances 
		SET check_in = $1, lunch_start = $2, lunch_end = $3, check_out = $4, 
		    is_absence = $5, absence_reason = $6, updated_at = NOW() 
		WHERE id = $7`,
		req.CheckIn, req.LunchStart, req.LunchEnd, req.CheckOut, 
		req.IsAbsence, req.AbsenceReason, id)

	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	// Re-evaluate incidents and recalculate earnings
	if err := h.AttendanceService.AutoDetectIncidents(h.DB, id); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	// Invalidate reports if the record was part of an existing period
	if oldAtt.CheckIn != nil {
		h.ReportService.InvalidateReports(oldAtt.EmployeeID, *oldAtt.CheckIn)
	}

	// 5. Log Action
	h.AuditService.LogAction(userID, models.AuditActionUpdateAttendance, "attendance", id, oldAtt, req, c.IP())

	return c.JSON(fiber.Map{"message": "Attendance updated and earnings recalculated successfully"})
}

// UpdateIncidentStatus allows admin to Approve or Justify an incident
func (h *AdminHandler) UpdateIncidentStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Status         string  `json:"status"`
		ResolutionNote *string `json:"resolution_note"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID := c.Locals("user_id").(int)
	idInt, _ := strconv.Atoi(id)

	// 0. Get OLD state for audit
	var oldIncident models.Incident
	h.DB.Get(&oldIncident, "SELECT * FROM incidents WHERE id = $1", idInt)

	var attID int
	err := h.DB.QueryRow(`
		UPDATE incidents 
		SET status = $1, resolution_note = $2, resolved_by = $3, updated_at = NOW() 
		WHERE id = $4 RETURNING attendance_id`,
		req.Status, req.ResolutionNote, userID, id).Scan(&attID)

	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	// Recalculate earnings for the associated attendance
	if err := h.AttendanceService.RecalculateAttendance(h.DB, attID); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	// Invalidate reports
	var att models.Attendance
	if err := h.DB.Get(&att, "SELECT * FROM attendances WHERE id = $1", attID); err == nil && att.CheckIn != nil {
		h.ReportService.InvalidateReports(att.EmployeeID, *att.CheckIn)
	}

	// 5. Log Action
	h.AuditService.LogAction(userID, models.AuditActionUpdateIncidentStatus, "incident", idInt, oldIncident, req, c.IP())

	return c.JSON(fiber.Map{"message": "Incident status updated and earnings recalculated"})
}

// GetComplianceDashboard returns real-time stats per work center
func (h *AdminHandler) GetComplianceDashboard(c *fiber.Ctx) error {
	type CenterCompliance struct {
		CenterID        int    `json:"center_id"`
		CenterName      string `json:"center_name"`
		TotalExpected   int    `json:"total_expected"`
		PresentCount    int    `json:"present_count"`
		LateCount       int    `json:"late_count"`
		OutOfRangeCount int    `json:"out_of_range_count"`
	}

	var stats []CenterCompliance
	query := `
		SELECT 
			wc.id as center_id, 
			wc.name as center_name,
			(SELECT COUNT(*) FROM employees e WHERE e.work_center_id = wc.id AND e.is_active = true) as total_expected,
			(SELECT COUNT(*) FROM attendances a WHERE a.work_center_id = wc.id AND a.check_in::date = CURRENT_DATE AND a.check_out IS NULL) as present_count,
			(SELECT COUNT(*) FROM incidents i WHERE i.work_center_id = wc.id AND i.created_at::date = CURRENT_DATE AND i.type = 'late') as late_count,
			(SELECT COUNT(*) FROM incidents i WHERE i.work_center_id = wc.id AND i.created_at::date = CURRENT_DATE AND i.type = 'out_of_range') as out_of_range_count
		FROM work_centers wc
		ORDER BY wc.name
	`
	
	if err := h.DB.Select(&stats, query); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	return c.JSON(stats)
}

// RecalculateIncidents triggers an automatic scan for infractions on an existing record
func (h *AdminHandler) RecalculateIncidents(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")

	if err := h.AttendanceService.AutoDetectIncidents(h.DB, id); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	// Invalidate reports
	var att models.Attendance
	if err := h.DB.Get(&att, "SELECT * FROM attendances WHERE id = $1", id); err == nil && att.CheckIn != nil {
		h.ReportService.InvalidateReports(att.EmployeeID, *att.CheckIn)
	}

	return c.JSON(fiber.Map{"message": "Incidencias recalculadas exitosamente"})
}

// AUDIT LOGS
func (h *AdminHandler) ListAuditLogs(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	if page < 1 { page = 1 }
	if limit < 1 { limit = 20 }
	offset := (page - 1) * limit

	action := c.Query("action")
	entity := c.Query("entity")
	start := c.Query("start")
	end := c.Query("end")

	where := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if (action != "" && action != "all") {
		where += fmt.Sprintf(" AND l.action ILIKE $%d", argCount)
		args = append(args, "%"+action+"%")
		argCount++
	}
	if (entity != "" && entity != "all") {
		where += fmt.Sprintf(" AND l.entity_type ILIKE $%d", argCount)
		args = append(args, "%"+entity+"%")
		argCount++
	}
	if start != "" {
		where += fmt.Sprintf(" AND l.created_at >= $%d", argCount)
		args = append(args, start)
		argCount++
	}
	if end != "" {
		where += fmt.Sprintf(" AND l.created_at <= $%d", argCount)
		args = append(args, end)
		argCount++
	}

	var total int
	h.DB.Get(&total, "SELECT COUNT(*) FROM audit_logs l " + where, args...)

	logs := []models.AuditLog{}
	query := fmt.Sprintf(`
		SELECT l.*, COALESCE(u.name, 'System') as user_name
		FROM audit_logs l
		LEFT JOIN users u ON l.user_id = u.id
		%s
		ORDER BY l.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argCount, argCount+1)
	
	args = append(args, limit, offset)

	if err := h.DB.Select(&logs, query, args...); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	totalPages := (total + limit - 1) / limit

	return c.JSON(fiber.Map{
		"data":        logs,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// SYSTEM ALERTS
func (h *AdminHandler) ListAlerts(c *fiber.Ctx) error {
	onlyUnread := c.Query("unread") == "true"
	limit := c.QueryInt("limit", 50)

	alerts, err := h.AlertService.ListAlerts(onlyUnread, limit)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	return c.JSON(alerts)
}

func (h *AdminHandler) MarkAlertAsRead(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	if err := h.AlertService.MarkAsRead(id); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// JUSTIFICATIONS
func (h *AdminHandler) ListJustifications(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)
	list, err := h.JustificationService.ListPending(limit)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}
	return c.JSON(list)
}

func (h *AdminHandler) ResolveJustification(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	var req struct {
		Approve bool   `json:"approve"`
		Note    string `json:"note"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	adminID := c.Locals("user_id").(int)
	if err := h.JustificationService.ResolveJustification(id, adminID, req.Approve, req.Note); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Justificación resuelta correctamente"})
}

// BULK ACTIONS
func (h *AdminHandler) BulkUpdateEmployees(c *fiber.Ctx) error {
	var req models.BulkEmployeeUpdate
	if err := utils.ParseAndValidate(c, &req); err != nil {
		return err
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Could not start transaction", err)
	}

	userID := c.Locals("user_id").(int)
	ip := c.IP()

	for _, id := range req.IDs {
		var old models.Employee
		if err := tx.Get(&old, "SELECT * FROM employees WHERE id = $1", id); err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusNotFound, fmt.Sprintf("Employee ID %d not found", id), err)
		}

		query := "UPDATE employees SET updated_at = NOW()"
		args := map[string]interface{}{"id": id}

		if req.WorkCenterID != nil {
			query += ", work_center_id = :wc"
			args["wc"] = *req.WorkCenterID
		}
		if req.WorkShiftID != nil {
			query += ", work_shift_id = :ws"
			args["ws"] = *req.WorkShiftID
		}
		if req.IsActive != nil {
			query += ", is_active = :active"
			args["active"] = *req.IsActive
		}
		query += " WHERE id = :id"

		if _, err := tx.NamedExec(query, args); err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update employee batch", err)
		}

		if err := h.AuditService.LogActionTx(tx, userID, models.AuditActionBulkUpdateEmployees, "employee", id, old, args, ip); err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Error logging batch action", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to commit transaction", err)
	}

	h.Cache.Delete("employees") // Invalidate employee cache if it exists
	return c.JSON(fiber.Map{"message": fmt.Sprintf("Successfully updated %d employees", len(req.IDs))})
}

func (h *AdminHandler) BulkDeleteEmployees(c *fiber.Ctx) error {
	var req models.BulkRequest
	if err := utils.ParseAndValidate(c, &req); err != nil {
		return err
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Could not start transaction", err)
	}

	userID := c.Locals("user_id").(int)
	ip := c.IP()

	for _, id := range req.IDs {
		var old models.Employee
		if err := tx.Get(&old, "SELECT * FROM employees WHERE id = $1", id); err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusNotFound, fmt.Sprintf("Employee ID %d not found", id), err)
		}

		if _, err := tx.Exec("DELETE FROM employees WHERE id = $1", id); err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusBadRequest, fmt.Sprintf("Cannot delete employee %d: likely has linked attendance records", id), err)
		}

		if err := h.AuditService.LogActionTx(tx, userID, models.AuditActionBulkDeleteEmployees, "employee", id, old, nil, ip); err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Error logging batch action", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to commit transaction", err)
	}

	h.Cache.Delete("employees")
	return c.JSON(fiber.Map{"message": fmt.Sprintf("Successfully deleted %d employees", len(req.IDs))})
}

func (h *AdminHandler) BulkJustifyAttendances(c *fiber.Ctx) error {
	var req models.BulkJustifyRequest
	if err := utils.ParseAndValidate(c, &req); err != nil {
		return err
	}

	adminID := c.Locals("user_id").(int)
	tx, err := h.DB.Beginx()
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Could not start transaction", err)
	}

	for _, id := range req.AttendanceIDs {
		if _, err := tx.Exec("UPDATE incidents SET status = $1, resolution_note = $2, resolved_by = $3 WHERE attendance_id = $4 AND status = $5", 
			models.StatusJustified, req.Note, adminID, id, models.StatusPending); err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update incidents", err)
		}

		if err := h.AttendanceService.AutoDetectIncidentsTx(tx, id); err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to recalculate totals", err)
		}

		h.AuditService.LogActionTx(tx, adminID, models.AuditActionBulkJustifyAttendances, "attendance", id, nil, req.Note, c.IP())
	}

	if err := tx.Commit(); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to commit transaction", err)
	}

	return c.JSON(fiber.Map{"message": fmt.Sprintf("Successfully justified %d attendance records", len(req.AttendanceIDs))})
}

// INCIDENTS
func (h *AdminHandler) ListIncidents(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	status := c.Query("status", "all")
	typeFilter := c.Query("type", "all")
	search := c.Query("search")
	start := c.Query("start")
	end := c.Query("end")
	shiftID := c.Query("shift_id", "all")
	centerID := c.Query("center_id", "all")
	positionID := c.Query("position_id", "all")

	offset := (page - 1) * limit

	where := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if status != "all" {
		where += fmt.Sprintf(" AND i.status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	if typeFilter != "all" {
		where += fmt.Sprintf(" AND i.type = $%d", argCount)
		args = append(args, typeFilter)
		argCount++
	}

	if search != "" {
		where += fmt.Sprintf(" AND u.name ILIKE $%d", argCount)
		args = append(args, "%"+search+"%")
		argCount++
	}

	if start != "" {
		where += fmt.Sprintf(" AND i.created_at >= $%d", argCount)
		args = append(args, start)
		argCount++
	}

	if end != "" {
		where += fmt.Sprintf(" AND i.created_at <= $%d", argCount)
		args = append(args, end)
		argCount++
	}

	if shiftID != "all" {
		where += fmt.Sprintf(" AND e.work_shift_id = $%d", argCount)
		args = append(args, shiftID)
		argCount++
	}

	if centerID != "all" {
		where += fmt.Sprintf(" AND i.work_center_id = $%d", argCount)
		args = append(args, centerID)
		argCount++
	}

	if positionID != "all" {
		where += fmt.Sprintf(" AND e.position_id = $%d", argCount)
		args = append(args, positionID)
		argCount++
	}

	shiftType := c.Query("shift_type")
	if shiftType != "" && shiftType != "all" {
		where += fmt.Sprintf(" AND ws.shift_type = $%d", argCount)
		args = append(args, shiftType)
		argCount++
	}

	// Security: Manager/Supervisor restriction
	roleSlug, _ := c.Locals("role_slug").(string)
	userID, _ := c.Locals("user_id").(int)
	if roleSlug != "" && roleSlug != "admin" {
		where += fmt.Sprintf(" AND i.work_center_id IN (SELECT id FROM work_centers WHERE manager_id = $%d)", argCount)
		args = append(args, userID)
		argCount++
	}

	type IncidentRecord struct {
		models.Incident
		EmployeeName   string `db:"employee_name"`
		AttendanceDate string `db:"attendance_date"`
		CenterName     string `db:"center_name"`
	}

	var records []IncidentRecord
	query := fmt.Sprintf(`
		SELECT i.*, u.name as employee_name, a.check_in::date::text as attendance_date, wc.name as center_name
		FROM incidents i
		JOIN employees e ON i.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		JOIN attendances a ON i.attendance_id = a.id
		JOIN work_centers wc ON i.work_center_id = wc.id
		%s
		ORDER BY i.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argCount, argCount+1)

	listArgs := append(args, limit, offset)
	if err := h.DB.Select(&records, query, listArgs...); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to fetch incidents", err)
	}

	var total int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM incidents i 
		JOIN employees e ON i.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		%s
	`, where)
	if err := h.DB.Get(&total, countQuery, args...); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to count incidents", err)
	}

	dtos := []models.IncidentRichDTO{}
	for _, r := range records {
		dtos = append(dtos, models.MapIncidentToRichDTO(r.Incident, r.EmployeeName, r.AttendanceDate, r.CenterName))
	}

	return c.JSON(models.PaginatedResponse{
		Data:       dtos,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: (total + limit - 1) / limit,
	})
}

func (h *AdminHandler) BulkJustifyIncidents(c *fiber.Ctx) error {
	var req models.BulkJustifyIncidentsRequest
	if err := utils.ParseAndValidate(c, &req); err != nil {
		return err
	}

	adminID := c.Locals("user_id").(int)
	tx, err := h.DB.Beginx()
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Could not start transaction", err)
	}

	// Track affected attendances for recalculation
	attendanceIDs := make(map[int]bool)

	for _, id := range req.IncidentIDs {
		var attID int
		err := tx.Get(&attID, "UPDATE incidents SET status = $1, resolution_note = $2, resolved_by = $3 WHERE id = $4 AND status = $5 RETURNING attendance_id",
			models.StatusJustified, req.Note, adminID, id, models.StatusPending)
		
		if err != nil {
			// Skip if not found or already resolved, but in a real scenario we might want to log this
			continue
		}
		attendanceIDs[attID] = true
	}

	for attID := range attendanceIDs {
		if err := h.AttendanceService.AutoDetectIncidentsTx(tx, attID); err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to recalculate attendance totals", err)
		}
		h.AuditService.LogActionTx(tx, adminID, models.AuditActionBulkJustifyAttendances, "attendance", attID, nil, "Bulk Incident Justification: "+req.Note, c.IP())
	}

	if err := tx.Commit(); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to commit transaction", err)
	}

	return c.JSON(fiber.Map{"message": fmt.Sprintf("Successfully justified %d incidents", len(req.IncidentIDs))})
}




