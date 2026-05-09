package handlers

import (
	"attendance-api/internal/models"
	"attendance-api/internal/services"
	"attendance-api/internal/utils"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	_ "time/tzdata"
)

type AttendanceHandler struct {
	DB                   *sqlx.DB
	Service              *services.AttendanceService
	JustificationService *services.JustificationService
}

type AttendanceRequest struct {
	EmployeeID   int     `json:"employee_id" validate:"required"`
	Latitude     float64 `json:"latitude" validate:"required"`
	Longitude    float64 `json:"longitude" validate:"required"`
	WorkCenterID *int    `json:"work_center_id"`
	EvidenceURL  *string `json:"evidence_url"`
	CheckOutNote *string `json:"check_out_note"`
	Address      *string  `json:"address"`
	CheckInNote  *string  `json:"check_in_note"`
	IsFieldWork  bool     `json:"is_field_work"`
	EvidenceURLs []string `json:"evidence_urls"`
}

type AbsenceRequest struct {
	EmployeeID int    `json:"employee_id" validate:"required"`
	Reason     string `json:"reason" validate:"required,min=5"`
}

// ListCenters returns all active work centers for employee selection
func (h *AttendanceHandler) ListCenters(c *fiber.Ctx) error {
	var centers []models.WorkCenter
	err := h.DB.Select(&centers, "SELECT * FROM work_centers ORDER BY name ASC")
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Error fetching work centers", err)
	}
	
	// Initialize slice to empty array if nil
	if centers == nil {
		centers = []models.WorkCenter{}
	}

	return c.JSON(centers)
}

// CheckIn registers check-in with GPS validation
func (h *AttendanceHandler) CheckIn(c *fiber.Ctx) error {
	var req AttendanceRequest
	if err := utils.ParseAndValidate(c, &req); err != nil {
		return err
	}

	// SECURITY & VALIDATION: Ensure profile exists, is active and complete
	userID := c.Locals("user_id").(int)
	var emp models.Employee
	err := h.DB.Get(&emp, "SELECT * FROM employees WHERE user_id = $1", userID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied or employee profile not found"})
	}

	if !emp.IsActive {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Acceso denegado: Tu cuenta de empleado está inactiva. Contacta a un administrador."})
	}

	if emp.WorkCenterID == 0 || emp.PositionID == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Perfil incompleto: Debes tener una sede y un puesto asignados para marcar asistencia."})
	}

	// 1. Determine WorkCenter (Use request override or employee default)
	centerID := emp.WorkCenterID
	if req.WorkCenterID != nil && *req.WorkCenterID != 0 {
		centerID = *req.WorkCenterID
	}

	var center models.WorkCenter
	if err := h.DB.Get(&center, "SELECT * FROM work_centers WHERE id = $1", centerID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Centro de trabajo no encontrado"})
	}

	// 2. Validate Geofence is now handled during insertion logic

	// 3. Create Attendance
	now := time.Now()
	
	// 1. Check for OPEN sessions (forgetting to check-out)
	// We only block if the open session belongs to the SAME calendar day.
	// If it's from yesterday, we allow the new check-in (the old one remains as "Incomplete" for admin review).
	loc, _ := time.LoadLocation(center.Timezone)
	if loc == nil { loc = time.UTC }
	nowLocal := now.In(loc)
	todayStr := nowLocal.Format("2006-01-02")

	var openAtt models.Attendance
	err = h.DB.Get(&openAtt, "SELECT * FROM attendances WHERE employee_id = $1 AND check_out IS NULL ORDER BY check_in DESC LIMIT 1", emp.ID)
	if err == nil && openAtt.CheckIn != nil {
		// If the open session is recent (e.g. < 16 hours), we block. 
		// If it's older, we assume it's a ghost session and allow a new check-in.
		if now.Sub(*openAtt.CheckIn) < 16*time.Hour {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Ya tienes una sesión de asistencia activa. Debes marcar salida antes de iniciar una nueva.",
				"attendance_id": openAtt.ID,
			})
		}
	}

	// 2. Prevent double entry for the SAME logical day -> REMOVED to support multi-checks
	// The open session check above already ensures they can't have two active sessions.

	// 3. Check for Mandatory Holidays
	var holiday models.Holiday
	err = h.DB.Get(&holiday, "SELECT * FROM holidays WHERE date::date = $1 AND type = $2 LIMIT 1", todayStr, models.HolidayTypeMandatory)
	if err == nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": fmt.Sprintf("Hoy es día festivo obligatorio (%s). No se permite el registro de asistencia.", holiday.Name),
		})
	}

	var shiftID *int
	isLate := false
	delayMinutes := 0
	var shift models.WorkShift
	
	if emp.WorkShiftID != 0 {
		shiftID = &emp.WorkShiftID

		// Validate Work Day
		if err := h.DB.Get(&shift, "SELECT * FROM work_shifts WHERE id = $1", emp.WorkShiftID); err == nil {
			var days []int
			if len(shift.WorkDays) > 0 {
				_ = json.Unmarshal(shift.WorkDays, &days)
			}

			today := int(now.Weekday())
			found := false
			for _, d := range days {
				if d == today {
					found = true
					break
				}
			}
			if !found && !req.IsFieldWork {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Hoy no es un día laboral programado para tu turno."})
			}

			// 2.2 Validate Time Windows (Only for fixed shifts, skipped if Field Work)
			if shift.ShiftType == "fixed" && !req.IsFieldWork {
				if h.Service.IsTooEarly(now, shift, center.Timezone) {
					startTime, _ := h.Service.ParseFlexibleTime(shift.ExpectedCheckIn)
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": fmt.Sprintf("Es demasiado pronto para marcar entrada. Tu turno inicia a las %s.", startTime.Format("15:04"))})
				}

				if h.Service.IsShiftOver(now, shift, center.Timezone) {
					overH, overM := h.Service.GetShiftOverDuration(now, shift, center.Timezone)
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": fmt.Sprintf("Tu turno finalizó hace %d horas y %d minutos.", overH, overM)})
				}
			}

			// 2.3 Check for Late Incident (Only if not flexible/field, skipped if Field Work)
			if shift.ShiftType == "fixed" && !req.IsFieldWork {
				isLate, delayMinutes = h.Service.CheckIfLate(now, shift, center.Timezone)
				utils.Logger.Info("Late Check", 
					zap.Time("now", now),
					zap.String("shift", shift.Name),
					zap.String("tz", center.Timezone),
					zap.Bool("isLate", isLate),
					zap.Int("delayMinutes", delayMinutes),
				)
			}
		}
	}

	attendance := models.Attendance{
		EmployeeID:       emp.ID,
		WorkCenterID:    &center.ID,
		WorkShiftID:      shiftID,
		CheckIn:          &now,
		CheckInLatitude:  &req.Latitude,
		CheckInLongitude: &req.Longitude,
		CheckInAddress:   req.Address,
		CheckInNote:      req.CheckInNote,
		IsFieldWork:      req.IsFieldWork,
		CreatedAt:        &now,
		UpdatedAt:        &now,
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	var newID int
	err = tx.QueryRow(
		`INSERT INTO attendances (employee_id, work_center_id, work_shift_id, check_in, check_in_latitude, check_in_longitude, check_in_address, check_in_note, is_field_work, created_at, updated_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`,
		emp.ID, center.ID, shiftID, now, req.Latitude, req.Longitude, req.Address, req.CheckInNote, req.IsFieldWork, now, now,
	).Scan(&newID)
	
	if err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}
	attendance.ID = newID

	// 4. Create Geofence Incident if necessary (Skip for field shifts OR explicit field work)
	if shift.ShiftType != "field" && !req.IsFieldWork {
		err = h.Service.CreateGeofenceIncident(c.Context(), tx, emp.ID, center.ID, attendance.ID, req.Latitude, req.Longitude, center, "check-in")
		if err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Error creating geofence incident", err)
		}
	}

	// 5. Create Late Incident if necessary
	if isLate && shift.EnforceLateness {
		_, err = tx.Exec(`INSERT INTO incidents 
			(employee_id, work_center_id, attendance_id, type, metadata_json, is_late, delay_minutes, is_out_of_range, distance, status, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`, 
			emp.ID, center.ID, attendance.ID, models.IncidentTypeLate, `{"action": "check-in", "status": "late"}`, true, delayMinutes, false, 0, models.StatusPending, now, now)
		if err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Error creating lateness incident", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	// Fetch updated attendance to return full DTO
	var updatedAtt models.Attendance
	h.DB.Get(&updatedAtt, "SELECT * FROM attendances WHERE id = $1", attendance.ID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "Check-in successful",
		"attendance": models.MapAttendanceToDTO(updatedAtt),
	})
}

// CheckOutNoID auto-finds active session
func (h *AttendanceHandler) CheckOutNoID(c *fiber.Ctx) error {
	var req AttendanceRequest
	if err := utils.ParseAndValidate(c, &req); err != nil { return err }

	// SECURITY & VALIDATION
	userID := c.Locals("user_id").(int)
	var emp models.Employee
	err := h.DB.Get(&emp, "SELECT * FROM employees WHERE user_id = $1", userID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	if !emp.IsActive {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Cuenta inactiva"})
	}

	// Find the most recent active session (check_out is NULL)
	var att models.Attendance
	if err := h.DB.Get(&att, "SELECT * FROM attendances WHERE employee_id = $1 AND check_out IS NULL ORDER BY check_in DESC LIMIT 1", emp.ID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No tienes una sesión de asistencia activa para cerrar."})
	}

	// MANDATORY EVIDENCE CHECK (Minimum 2)
	if len(req.EvidenceURLs) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Se requieren al menos 2 evidencias fotográficas para marcar salida."})
	}

	var center models.WorkCenter
	h.DB.Get(&center, "SELECT * FROM work_centers WHERE id = $1", att.WorkCenterID)

	var shift models.WorkShift
	if att.WorkShiftID != nil {
		h.DB.Get(&shift, "SELECT * FROM work_shifts WHERE id = $1", *att.WorkShiftID)
	}

	now := time.Now()
	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	// 1. Geofence Validation for Check-Out (Skip for field shifts OR field work)
	if shift.ShiftType != "field" && !att.IsFieldWork {
		err = h.Service.CreateGeofenceIncident(c.Context(), tx, emp.ID, center.ID, att.ID, req.Latitude, req.Longitude, center, "check-out")
		if err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Error creating checkout incident", err)
		}
	}

	// 2. Mark checkout in DB
	evsJSON, _ := json.Marshal(req.EvidenceURLs)
	
	// Use the first evidence as primary URL for legacy compatibility
	primaryURL := ""
	if len(req.EvidenceURLs) > 0 { primaryURL = req.EvidenceURLs[0] }

	_, err = tx.Exec(`UPDATE attendances SET check_out = $1, check_out_latitude = $2, check_out_longitude = $3, evidence_url = $4, check_out_note = $5, check_out_address = $6, evidence_urls = $7, updated_at = $1 WHERE id = $8`, 
		now, req.Latitude, req.Longitude, primaryURL, req.CheckOutNote, req.Address, evsJSON, att.ID)
	
	if err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	// 3. Auto-detect and Recalculate (Transactional)
	if err := h.Service.AutoDetectIncidents(c.Context(), tx, att.ID); err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Error calculating session totals", err)
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	// Fetch updated attendance to return DTO
	var updatedAtt models.Attendance
	h.DB.Get(&updatedAtt, "SELECT * FROM attendances WHERE id = $1", att.ID)

	return c.JSON(fiber.Map{
		"message":    "Check-out successful",
		"hours":      updatedAtt.NetHoursWorked,
		"earnings":   updatedAtt.DailyEarnings,
		"attendance": models.MapAttendanceToDTO(updatedAtt),
	})
}

// CheckOut by specific ID (Fortified for owners)
func (h *AttendanceHandler) CheckOut(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(int)
	
	var att models.Attendance
	if err := h.DB.Get(&att, "SELECT * FROM attendances WHERE id = $1", id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Attendance not found"})
	}

	// Ownership check
	var count int
	h.DB.Get(&count, "SELECT COUNT(*) FROM employees WHERE id = $1 AND user_id = $2", att.EmployeeID, userID)
	if count == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	now := time.Now()

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	var center models.WorkCenter
	if err := tx.Get(&center, "SELECT * FROM work_centers WHERE id = $1", att.WorkCenterID); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Work center not found"})
	}

	// We need coordinates from the request, but the standard CheckOut(id) doesn't always have them if called as a simple DELETE/PUT
	// HOWEVER, the AttendanceRequest struct has them. We should expect them here too for geofencing.
	var req AttendanceRequest
	_ = utils.ParseAndValidate(c, &req) // Optional coordinates, but if provided, must be valid
	var shift models.WorkShift
	if att.WorkShiftID != nil {
		tx.Get(&shift, "SELECT * FROM work_shifts WHERE id = $1", *att.WorkShiftID)
	}

	// Geofence Validation (Skip for field shifts OR field work)
	if req.Latitude != 0 && req.Longitude != 0 && shift.ShiftType != "field" && !att.IsFieldWork {
		err = h.Service.CreateGeofenceIncident(c.Context(), tx, att.EmployeeID, center.ID, att.ID, req.Latitude, req.Longitude, center, "check-out-manual")
		if err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Error creating geofence incident", err)
		}
	}

	// MANDATORY EVIDENCE CHECK (Minimum 2)
	if len(req.EvidenceURLs) < 2 {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Se requieren al menos 2 evidencias fotográficas para marcar salida."})
	}

	// 3. Mark checkout in DB
	evsJSON, _ := json.Marshal(req.EvidenceURLs)
	primaryURL := ""
	if len(req.EvidenceURLs) > 0 { primaryURL = req.EvidenceURLs[0] }

	_, err = tx.Exec("UPDATE attendances SET check_out = $1, check_out_latitude = $2, check_out_longitude = $3, evidence_url = $4, check_out_note = $5, check_out_address = $6, evidence_urls = $7, updated_at = $1 WHERE id = $8", 
		now, req.Latitude, req.Longitude, primaryURL, req.CheckOutNote, req.Address, evsJSON, att.ID)
	if err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	// 4. Auto-detect and Recalculate everything (Transactional)
	if err := h.Service.AutoDetectIncidents(c.Context(), tx, att.ID); err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Error calculating session totals", err)
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	// Fetch updated to return DTO
	var updatedAtt models.Attendance
	h.DB.Get(&updatedAtt, "SELECT * FROM attendances WHERE id = $1", att.ID)

	return c.JSON(fiber.Map{
		"message":    "Check-out successful",
		"hours":      updatedAtt.NetHoursWorked,
		"earnings":   updatedAtt.DailyEarnings,
		"attendance": models.MapAttendanceToDTO(updatedAtt),
	})
}

func (h *AttendanceHandler) LunchStart(c *fiber.Ctx) error {
	var req AttendanceRequest
	if err := utils.ParseAndValidate(c, &req); err != nil { return err }

	userID := c.Locals("user_id").(int)
	var emp models.Employee
	err := h.DB.Get(&emp, "SELECT * FROM employees WHERE user_id = $1", userID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	if !emp.IsActive {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Cuenta inactiva"})
	}

	if emp.WorkCenterID == 0 || emp.PositionID == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Perfil incompleto"})
	}

	// Find the most recent active session
	var att models.Attendance
	if err := h.DB.Get(&att, "SELECT * FROM attendances WHERE employee_id = $1 AND check_out IS NULL ORDER BY check_in DESC LIMIT 1", emp.ID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No tienes una sesión activa para iniciar el almuerzo."})
	}

	now := time.Now()

	var center models.WorkCenter
	h.DB.Get(&center, "SELECT * FROM work_centers WHERE id = $1", att.WorkCenterID)
	// Geofence handled via service

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	_, err = tx.Exec(`UPDATE attendances SET lunch_start = $1, updated_at = $2 WHERE id = $3 AND lunch_start IS NULL`, 
		now, now, att.ID)
	
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not register lunch start"})
	}

	var shift models.WorkShift
	if att.WorkShiftID != nil {
		tx.Get(&shift, "SELECT * FROM work_shifts WHERE id = $1", *att.WorkShiftID)
	}

	// 3. Create Geofence Incident if necessary (Skip for field shifts OR field work)
	if shift.ShiftType != "field" && !att.IsFieldWork {
		err = h.Service.CreateGeofenceIncident(c.Context(), tx, emp.ID, center.ID, att.ID, req.Latitude, req.Longitude, center, "lunch-start")
		if err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Error creating lunch incident", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	// Fetch updated attendance
	var updatedAtt models.Attendance
	h.DB.Get(&updatedAtt, "SELECT * FROM attendances WHERE id = $1", att.ID)

	return c.JSON(fiber.Map{
		"message":    "Lunch started",
		"attendance": models.MapAttendanceToDTO(updatedAtt),
	})
}

func (h *AttendanceHandler) LunchEnd(c *fiber.Ctx) error {
	var req AttendanceRequest
	if err := utils.ParseAndValidate(c, &req); err != nil { return err }

	userID := c.Locals("user_id").(int)
	var emp models.Employee
	err := h.DB.Get(&emp, "SELECT * FROM employees WHERE user_id = $1", userID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	if !emp.IsActive {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Cuenta inactiva"})
	}

	// Find the most recent active session
	var att models.Attendance
	if err := h.DB.Get(&att, "SELECT * FROM attendances WHERE employee_id = $1 AND check_out IS NULL ORDER BY check_in DESC LIMIT 1", emp.ID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No tienes una sesión activa."})
	}

	now := time.Now()

	var center models.WorkCenter
	h.DB.Get(&center, "SELECT * FROM work_centers WHERE id = $1", att.WorkCenterID)
	// Geofence handled via service

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	_, err = tx.Exec(`UPDATE attendances SET lunch_end = $1, updated_at = $2 WHERE id = $3 AND lunch_end IS NULL AND lunch_start IS NOT NULL`, 
		now, now, att.ID)
	
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not register lunch end"})
	}

	var shift models.WorkShift
	if att.WorkShiftID != nil {
		tx.Get(&shift, "SELECT * FROM work_shifts WHERE id = $1", *att.WorkShiftID)
	}

	// 3. Create Geofence Incident if necessary (Skip for field shifts OR field work)
	if shift.ShiftType != "field" && !att.IsFieldWork {
		err = h.Service.CreateGeofenceIncident(c.Context(), tx, emp.ID, center.ID, att.ID, req.Latitude, req.Longitude, center, "lunch-end")
		if err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Error creating lunch incident", err)
		}
	}

	// 4. Auto-detect incidents (Lunch Overstay) and Recalculate financial data
	if err := h.Service.AutoDetectIncidents(c.Context(), tx, att.ID); err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Error recalculating attendance", err)
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	// Fetch updated attendance
	var updatedAtt models.Attendance
	h.DB.Get(&updatedAtt, "SELECT * FROM attendances WHERE id = $1", att.ID)

	return c.JSON(fiber.Map{
		"message":    "Lunch ended",
		"attendance": models.MapAttendanceToDTO(updatedAtt),
	})
}

func (h *AttendanceHandler) ReportAbsence(c *fiber.Ctx) error {
	var req AbsenceRequest
	if err := utils.ParseAndValidate(c, &req); err != nil {
		return err
	}

	if req.Reason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Se requiere un motivo para la ausencia"})
	}

	userID := c.Locals("user_id").(int)
	var emp models.Employee
	err := h.DB.Get(&emp, "SELECT * FROM employees WHERE user_id = $1", userID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	// Get WorkCenter for Timezone
	var center models.WorkCenter
	if err := h.DB.Get(&center, "SELECT * FROM work_centers WHERE id = $1", emp.WorkCenterID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Work center not assigned or not found"})
	}

	loc, _ := time.LoadLocation(center.Timezone)
	if loc == nil { loc = time.UTC }
	now := time.Now()
	nowLocal := now.In(loc)
	dateStr := nowLocal.Format("2006-01-02")

	// Check if any attendance record exists for today (check-in or absence)
	var existingCount int
	// For absences, we check either check_in or created_at (localized)
	h.DB.Get(&existingCount, "SELECT COUNT(*) FROM attendances WHERE employee_id = $1 AND (check_in::date = $2 OR created_at::date = $2)", emp.ID, dateStr)
	if existingCount > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Ya existe un registro de actividad o ausencia para hoy"})
	}

	var shiftID *int
	if emp.WorkShiftID != 0 {
		shiftID = &emp.WorkShiftID
	}


	attendance := models.Attendance{
		EmployeeID:    emp.ID,
		IsAbsence:     true,
		AbsenceReason: &req.Reason,
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	var newID int
	err = tx.QueryRow(
		`INSERT INTO attendances (employee_id, work_shift_id, is_absence, absence_reason, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		emp.ID, shiftID, true, req.Reason, now, now,
	).Scan(&newID)
	
	if err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}
	attendance.ID = newID

	// 4. Create Incident for absence
	_, err = tx.Exec(`INSERT INTO incidents 
		(employee_id, work_center_id, attendance_id, type, description, status, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`, 
		emp.ID, emp.WorkCenterID, attendance.ID, models.IncidentTypeAbsent, req.Reason, models.StatusPending, now, now)

	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error creating absence incident"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	// Fetch updated attendance
	var updatedAtt models.Attendance
	h.DB.Get(&updatedAtt, "SELECT * FROM attendances WHERE id = $1", attendance.ID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "Absence reported successfully",
		"attendance": models.MapAttendanceToDTO(updatedAtt),
	})
}

func (h *AttendanceHandler) GetTodayStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	var employee models.Employee
	err := h.DB.Get(&employee, "SELECT * FROM employees WHERE user_id = $1", userID)
	if err != nil {
		return c.JSON(fiber.Map{
			"is_employee": false,
			"attendance":  nil,
		})
	}

	var attendance models.Attendance
	// Query for any open session OR the most recent one within a 14h window 
	// (Night shifts usually don't exceed 14h total duration including overtime)
	err = h.DB.Get(&attendance, `
		SELECT * FROM attendances 
		WHERE employee_id = $1 
		AND (
			check_out IS NULL 
			OR (check_in > (NOW() - INTERVAL '14 hours'))
		)
		ORDER BY created_at DESC 
		LIMIT 1`, employee.ID)

	if err != nil {
		if err.Error() != "sql: no rows in result set" {
			utils.GetLogger().Error("GetTodayStatus Error", zap.Int("user_id", userID), zap.Int("employee_id", employee.ID), zap.Error(err))
		}
		return c.JSON(fiber.Map{
			"is_employee":   true,
			"employee_id":   employee.ID,
			"is_active":     employee.IsActive,
			"is_incomplete": employee.WorkCenterID == 0 || employee.PositionID == 0,
			"attendance":    nil,
		})
	}

	incompleteReason := ""
	if employee.WorkCenterID == 0 {
		incompleteReason = "Sede no asignada"
	} else if employee.PositionID == 0 {
		incompleteReason = "Puesto no asignado"
	}

	var shift models.WorkShift
	if employee.WorkShiftID != 0 {
		h.DB.Get(&shift, "SELECT * FROM work_shifts WHERE id = $1", employee.WorkShiftID)
	}

	var center models.WorkCenter
	if employee.WorkCenterID != 0 {
		h.DB.Get(&center, "SELECT * FROM work_centers WHERE id = $1", employee.WorkCenterID)
	}

	// Calculate Dynamic Goal on Server
	goalSeconds := 8.0 * 3600.0
	if shift.ID != 0 {
		// 1. Get Standard Duration (Independent of center timezone, as these are "wall clock" times)
		startTime, _ := h.Service.ParseFlexibleTime(shift.ExpectedCheckIn)
		endTime, _ := h.Service.ParseFlexibleTime(shift.ExpectedCheckOut)
		lunchTime, _ := h.Service.ParseFlexibleTime(shift.AllowedLunchTime)

		sh, sm, ss := startTime.Clock()
		eh, em, es := endTime.Clock()
		lh, lm, ls := lunchTime.Clock()

		startSec := float64(sh*3600 + sm*60 + ss)
		endSec := float64(eh*3600 + em*60 + es)
		lunchSec := float64(lh*3600 + lm*60 + ls)

		if shift.IsNightShift && endSec < startSec {
			endSec += 24 * 3600
		}

		standardDuration := endSec - startSec - lunchSec

		// 2. Adjust by Delay if Checked In
		if attendance.CheckIn != nil && center.ID != 0 {
			timezone := center.Timezone
			if timezone == "" { timezone = "America/Mexico_City" }
			loc, err := time.LoadLocation(timezone)
			if err != nil {
				utils.GetLogger().Error("Failed to load location", zap.String("timezone", timezone), zap.Error(err))
				loc = time.UTC
			}
			
			localCheckIn := attendance.CheckIn.In(loc)
			ch, cm, cs := localCheckIn.Clock()
			checkInSec := float64(ch*3600 + cm*60 + cs)
			
			// If night shift and check in is early morning (after midnight)
			if shift.IsNightShift && checkInSec < 12*3600 {
				checkInSec += 24 * 3600
			}

			delay := checkInSec - startSec
			
			// If we arrived before shift start, delay is negative, which INCREASES goalSeconds.
			// This is correct because if you arrive early, you are expected to stay until end_time.
			goalSeconds = standardDuration - delay
		} else {
			goalSeconds = standardDuration
		}
	}

	// Safety: Ensure goal is at least 0
	if goalSeconds < 0 {
		goalSeconds = 0
	}

	// Fetch ALL attendances for TODAY to support multi-check timeline
	var history []models.Attendance
	h.DB.Select(&history, `
		SELECT * FROM attendances 
		WHERE employee_id = $1 
		AND check_in::date = CURRENT_DATE
		ORDER BY check_in DESC
	`, employee.ID)

	historyDTOs := make([]models.AttendanceDTO, 0)
	for _, h_item := range history {
		historyDTOs = append(historyDTOs, models.MapAttendanceToDTO(h_item))
	}

	return c.JSON(fiber.Map{
		"is_employee":       true,
		"employee_id":       employee.ID,
		"is_active":         employee.IsActive,
		"is_incomplete":     employee.WorkCenterID == 0 || employee.PositionID == 0,
		"incomplete_reason": incompleteReason,
		"attendance":        models.MapAttendanceToDTO(attendance),
		"history":           historyDTOs,
		"shift":             shift,
		"center":            center,
		"goal_seconds":      goalSeconds,
	})
}

func (h *AttendanceHandler) SubmitJustification(c *fiber.Ctx) error {
	var req struct {
		IncidentID  int     `json:"incident_id"`
		Message     string  `json:"message"`
		EvidenceURL *string `json:"evidence_url"`
	}
	if err := utils.ParseAndValidate(c, &req); err != nil { return err }

	userID := c.Locals("user_id").(int)
	var emp models.Employee
	if err := h.DB.Get(&emp, "SELECT * FROM employees WHERE user_id = $1", userID); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Employee profile not found"})
	}

	if err := h.JustificationService.CreateJustification(c.Context(), nil, req.IncidentID, emp.ID, req.Message, req.EvidenceURL); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Justificación enviada correctamente"})
}






