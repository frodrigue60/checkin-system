package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"attendance-api/internal/models"
	"attendance-api/internal/utils"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (h *AdminBase) buildAttendanceFilters(c *fiber.Ctx) ([]string, []interface{}) {
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
func (h *AttendanceAdminHandler) ListAttendances(c *fiber.Ctx) error {
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
		dtos = append(dtos, models.MapAttendanceToDetailDTO(a.Attendance, a.EmployeeName, a.WorkCenterName, a.PositionName, a.IsLate, h.Cfg.R2PublicURL))
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


func (h *AttendanceAdminHandler) DeleteAttendance(c *fiber.Ctx) error {
	id := c.Params("id")

	userID := c.Locals("user_id").(int)
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{Code: models.ErrInvalidID})
	}

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
		h.ReportService.InvalidateReports(c.Context(), h.DB, oldAtt.EmployeeID, *oldAtt.CheckIn)
	}

	// 4. Log Action
	h.AuditService.LogAction(c.Context(), h.DB, userID, models.AuditActionDeleteAttendance, "attendance", idInt, oldAtt, nil, c.IP())

	return c.SendStatus(fiber.StatusNoContent)
}

// DASHBOARD STATS

func (h *AttendanceAdminHandler) GetAttendanceDetails(c *fiber.Ctx) error {
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
		utils.GetLogger().Error("Error fetching incidents for att", zap.String("id", id), zap.Error(err))
	}

	incidentDTOs := make([]models.IncidentDTO, 0)
	for _, i := range incidents {
		incidentDTOs = append(incidentDTOs, models.MapIncidentToDTO(i, h.Cfg.R2PublicURL))
	}

	return c.JSON(models.AttendanceFullDetailDTO{
		Attendance: models.MapAttendanceToDetailDTO(attRich.Attendance, attRich.EmployeeName, attRich.WorkCenterName, attRich.PositionName, attRich.IsLate, h.Cfg.R2PublicURL),
		WorkCenter: models.MapWorkCenterToDTO(center),
		Shift:      shiftDTO,
		Incidents:  incidentDTOs,
	})
}

// UpdateAttendance allows manual override of timestamps by admin
func (h *AttendanceAdminHandler) UpdateAttendance(c *fiber.Ctx) error {
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

	tx, err := h.DB.BeginTxx(c.Context(), nil)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to start transaction", err)
	}
	defer tx.Rollback()

	// Update timestamps and absence status
	_, err = tx.ExecContext(c.Context(), `
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
	if err := h.AttendanceService.AutoDetectIncidents(c.Context(), tx, id); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	if err := tx.Commit(); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to commit transaction", err)
	}

	// Invalidate reports if the record was part of an existing period
	if oldAtt.CheckIn != nil {
		h.ReportService.InvalidateReports(c.Context(), h.DB, oldAtt.EmployeeID, *oldAtt.CheckIn)
	}

	// 5. Log Action
	h.AuditService.LogAction(c.Context(), h.DB, userID, models.AuditActionUpdateAttendance, "attendance", id, oldAtt, req, c.IP())

	return c.JSON(fiber.Map{"message": "Attendance updated and earnings recalculated successfully"})
}

// UpdateIncidentStatus allows admin to Approve or Justify an incident

func (h *AttendanceAdminHandler) RecalculateIncidents(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")

	if err := h.AttendanceService.AutoDetectIncidents(c.Context(), h.DB, id); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	// Invalidate reports
	var att models.Attendance
	if err := h.DB.Get(&att, "SELECT * FROM attendances WHERE id = $1", id); err == nil && att.CheckIn != nil {
		h.ReportService.InvalidateReports(c.Context(), h.DB, att.EmployeeID, *att.CheckIn)
	}

	return c.JSON(fiber.Map{"message": "Incidencias recalculadas exitosamente"})
}

// AUDIT LOGS

