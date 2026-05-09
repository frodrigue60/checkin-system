package handlers

import (
	"fmt"
	"strconv"

	"attendance-api/internal/models"
	"attendance-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func (h *IncidentHandler) UpdateIncidentStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Status         string  `json:"status"`
		ResolutionNote *string `json:"resolution_note"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID := c.Locals("user_id").(int)
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{Code: models.ErrInvalidID})
	}

	// 0. Get OLD state for audit
	var oldIncident models.Incident
	h.DB.Get(&oldIncident, "SELECT * FROM incidents WHERE id = $1", idInt)

	tx, err := h.DB.BeginTxx(c.Context(), nil)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to start transaction", err)
	}
	defer tx.Rollback()

	var attID int
	err = tx.QueryRowxContext(c.Context(), `
		UPDATE incidents 
		SET status = $1, resolution_note = $2, resolved_by = $3, updated_at = NOW() 
		WHERE id = $4 RETURNING attendance_id`,
		req.Status, req.ResolutionNote, userID, id).Scan(&attID)

	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	// Recalculate Attendance Earnings if it belongs to an attendance
	if attID != 0 {
		if err := h.AttendanceService.RecalculateAttendance(c.Context(), tx, attID); err != nil {
			return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to commit transaction", err)
	}

	// Invalidate reports
	if attID != 0 {
		var att models.Attendance
		if err := h.DB.Get(&att, "SELECT * FROM attendances WHERE id = $1", attID); err == nil && att.CheckIn != nil {
			h.ReportService.InvalidateReports(c.Context(), h.DB, att.EmployeeID, *att.CheckIn)
		}
	}

	// 6. Log Action
	h.AuditService.LogAction(c.Context(), h.DB, userID, models.AuditActionUpdateIncidentStatus, "incident", idInt, oldIncident, req, c.IP())

	return c.JSON(fiber.Map{"message": "Incident status updated and earnings recalculated"})
}

// GetComplianceDashboard returns real-time stats per work center

func (h *IncidentHandler) ListIncidents(c *fiber.Ctx) error {
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
		dtos = append(dtos, models.MapIncidentToRichDTO(r.Incident, r.EmployeeName, r.AttendanceDate, r.CenterName, h.Cfg.R2PublicURL))
	}

	return c.JSON(models.PaginatedResponse{
		Data:       dtos,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: (total + limit - 1) / limit,
	})
}


