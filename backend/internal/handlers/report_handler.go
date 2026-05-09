package handlers

import (
	"attendance-api/internal/config"
	"attendance-api/internal/models"
	"attendance-api/internal/services"
	"attendance-api/internal/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type ReportHandler struct {
	DB           *sqlx.DB
	Cfg          *config.Config
	PDFService   *services.PDFService
	ExcelService *services.ExcelService
	Storage      services.StorageService
	AuditService *services.AuditService
}

type GenerateReportRequest struct {
	StartDate    string `json:"start_date" validate:"required"`
	EndDate      string `json:"end_date" validate:"required"`
	EmployeeID   *int   `json:"employee_id,omitempty"`
	WorkShiftID  *int   `json:"work_shift_id,omitempty"`
	PositionID   *int   `json:"position_id,omitempty"`
	WorkCenterID *int   `json:"work_center_id,omitempty"`
}

// List all unique report ranges generated
func (h *ReportHandler) ListReports(c *fiber.Ctx) error {
	type ReportRange struct {
		StartDate     time.Time `db:"start_date" json:"start_date"`
		EndDate       time.Time `db:"end_date" json:"end_date"`
		EmployeeCount int       `db:"emp_count" json:"employee_count"`
		IsStale       bool      `db:"is_stale" json:"is_stale"`
	}

	query := `
		SELECT 
			start_date, 
			end_date, 
			COUNT(employee_id) as emp_count,
			EXISTS (SELECT 1 FROM reports r2 WHERE r2.start_date = reports.start_date AND r2.end_date = reports.end_date AND r2.status = 'stale') as is_stale
		FROM reports 
		GROUP BY start_date, end_date 
		ORDER BY start_date DESC
	`
	ranges := []ReportRange{}
	err := h.DB.Select(&ranges, query)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	return c.JSON(ranges)
}

// Get details for a specific report range
func (h *ReportHandler) GetReportDetails(c *fiber.Ctx) error {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing start_date or end_date query params"})
	}

	query := `
		SELECT r.*, u.name as employee_name
		FROM reports r
		JOIN employees e ON r.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		WHERE r.start_date = $1 AND r.end_date = $2
	`
	var entities []models.Report
	err := h.DB.Select(&entities, query, startDate, endDate)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	// Map to DTOs
	dtos := make([]models.ReportDTO, 0)
	for _, r := range entities {
		dtos = append(dtos, models.MapReportToDTO(r))
	}

	return c.JSON(dtos)
}

// DownloadIndividualReport handles downloading a single report as PDF
func (h *ReportHandler) DownloadIndividualReport(c *fiber.Ctx) error {
	id := c.Params("id")

	var report models.Report
	err := h.DB.Get(&report, `
		SELECT r.*, u.name as employee_name 
		FROM reports r
		JOIN employees e ON r.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		WHERE r.id = $1
	`, id)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Reporte no encontrado"})
	}

	dto := models.MapReportToDTO(report)

	orientation := c.Query("orientation", "vertical")
	lang := c.Query("lang", "es")
	pdfBytes, err := h.PDFService.GenerateReportsPDF([]models.ReportDTO{dto}, orientation, lang)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al generar PDF"})
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf("inline; filename=reporte_%s.pdf", id))
	return c.Send(pdfBytes)
}

// DownloadBatchReport handles downloading multiple reports as a single PDF
func (h *ReportHandler) DownloadBatchReport(c *fiber.Ctx) error {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		return c.Status(400).JSON(fiber.Map{"error": "start_date y end_date son requeridos"})
	}

	var reports []models.Report
	err := h.DB.Select(&reports, `
		SELECT r.*, u.name as employee_name 
		FROM reports r
		JOIN employees e ON r.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		WHERE r.start_date = $1 AND r.end_date = $2
		ORDER BY u.name ASC
	`, startDate, endDate)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al consultar reportes"})
	}

	if len(reports) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "No se encontraron reportes para este período"})
	}

	var dtos []models.ReportDTO
	for _, r := range reports {
		dtos = append(dtos, models.MapReportToDTO(r))
	}

	orientation := c.Query("orientation", "vertical")
	lang := c.Query("lang", "es")
	pdfBytes, err := h.PDFService.GenerateReportsPDF(dtos, orientation, lang)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al generar PDF"})
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf("inline; filename=reporte_batch_%s_%s.pdf", startDate, endDate))
	return c.Send(pdfBytes)
}

// ListReportJobs returns a list of background report generation jobs
func (h *ReportHandler) ListReportJobs(c *fiber.Ctx) error {
	var jobs []models.ReportJob
	err := h.DB.Select(&jobs, "SELECT * FROM report_jobs ORDER BY created_at DESC LIMIT 50")
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	dtos := make([]models.ReportJobDTO, 0)
	for _, j := range jobs {
		dtos = append(dtos, models.MapReportJobToDTO(j, h.Cfg.R2PublicURL))
	}
	return c.JSON(dtos)
}

// GetReportJob returns the status of a specific job
func (h *ReportHandler) GetReportJob(c *fiber.Ctx) error {
	id := c.Params("id")
	var job models.ReportJob
	err := h.DB.Get(&job, "SELECT * FROM report_jobs WHERE id = $1", id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Job not found"})
	}

	return c.JSON(models.MapReportJobToDTO(job, h.Cfg.R2PublicURL))
}

// GenerateFullReport refactorizado para ser ASÍNCRONO y TRANSACCIONAL
func (h *ReportHandler) GenerateFullReport(c *fiber.Ctx) error {
	var req GenerateReportRequest
	if err := utils.ParseAndValidate(c, &req); err != nil {
		return err
	}

	userID := c.Locals("user_id").(int)
	start, _ := time.Parse("2006-01-02", req.StartDate)
	end, _ := time.Parse("2006-01-02", req.EndDate)

	// Iniciar transacción para la creación del job y auditoría inicial
	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al iniciar transacción"})
	}
	defer tx.Rollback()

	// 1. Crear el Job en la DB
	var jobID int
	err = tx.Get(&jobID, `
		INSERT INTO report_jobs (user_id, start_date, end_date, status, progress) 
		VALUES ($1, $2, $3, 'pending', 0) RETURNING id
	`, userID, start, end)

	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	// 2. Auditoría inicial dentro de la misma transacción
	err = h.AuditService.LogAction(c.Context(), tx, userID, models.AuditActionGenerateReport, "report", 0, req, fiber.Map{"job_id": jobID}, c.IP())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al registrar auditoría"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al confirmar creación de job"})
	}

	// 3. Lanzar Goroutine
	go h.processReportJob(jobID, req)

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Generación de reporte iniciada en segundo plano",
		"job_id":  jobID,
	})
}

// processReportJob contiene la lógica pesada, ahora con transacciones por empleado
func (h *ReportHandler) processReportJob(jobID int, req GenerateReportRequest) {
	// Marcar como procesando
	h.DB.Exec("UPDATE report_jobs SET status = 'processing', updated_at = NOW() WHERE id = $1", jobID)

	start, _ := time.Parse("2006-01-02", req.StartDate)
	end, _ := time.Parse("2006-01-02", req.EndDate)

	// Lógica de carga de datos
	type EmpData struct {
		models.Employee
		EmployeeName      string  `db:"employee_name"`
		BasePay           float64 `db:"base_pay"`
		LatePenalty       float64 `db:"late_penalty"`
		OutOfRangePenalty float64 `db:"out_of_range_penalty"`
		LunchPenalty      float64 `db:"lunch_penalty"`
		WorkCenterName    string  `db:"work_center_name"`
	}

	var employees []EmpData
	query := `
		SELECT e.*, u.name as employee_name, p.hourly_rate as base_pay, 
		       p.late_penalty_fee as late_penalty, p.out_of_range_fee as out_of_range_penalty,
		       p.lunch_overstay_fee as lunch_penalty,
		       wc.name as work_center_name
		FROM employees e
		JOIN users u ON e.user_id = u.id
		JOIN positions p ON e.position_id = p.id
		JOIN work_centers wc ON e.work_center_id = wc.id
		WHERE e.is_active = true
	`
	args := []interface{}{}
	if req.EmployeeID != nil {
		query += " AND e.id = ?"
		args = append(args, *req.EmployeeID)
	}
	if req.WorkShiftID != nil && *req.WorkShiftID != 0 {
		query += " AND e.work_shift_id = ?"
		args = append(args, *req.WorkShiftID)
	}
	if req.PositionID != nil && *req.PositionID != 0 {
		query += " AND e.position_id = ?"
		args = append(args, *req.PositionID)
	}
	if req.WorkCenterID != nil && *req.WorkCenterID != 0 {
		query += " AND e.work_center_id = ?"
		args = append(args, *req.WorkCenterID)
	}

	query = h.DB.Rebind(query)
	err := h.DB.Select(&employees, query, args...)
	if err != nil {
		errMsg := "Error fetching employees: " + err.Error()
		h.DB.Exec("UPDATE report_jobs SET status = 'failed', error_message = $1, updated_at = NOW() WHERE id = $2", errMsg, jobID)
		return
	}

	totalEmps := len(employees)
	h.DB.Exec("UPDATE report_jobs SET total_records = $1 WHERE id = $2", totalEmps, jobID)

	if totalEmps == 0 {
		h.DB.Exec("UPDATE report_jobs SET status = 'completed', progress = 100, updated_at = NOW() WHERE id = $1", jobID)
		return
	}

	// Carga masiva de feriados
	var holidays []models.Holiday
	h.DB.Select(&holidays, "SELECT * FROM holidays WHERE date BETWEEN $1 AND $2", start, end)
	holidayMap := make(map[string]models.Holiday)
	for _, hol := range holidays {
		holidayMap[hol.Date.Format("2006-01-02")] = hol
	}

	// Carga masiva de asistencias
	empIDs := make([]int, totalEmps)
	for i, emp := range employees {
		empIDs[i] = emp.ID
	}

	queryAtt, argsAtt, _ := sqlx.In(`
		SELECT * FROM attendances 
		WHERE employee_id IN (?) AND (
			(check_in IS NOT NULL AND check_in::date BETWEEN ? AND ?) OR
			(check_in IS NULL AND created_at::date BETWEEN ? AND ?)
		)
	`, empIDs, start, end, start, end)
	queryAtt = h.DB.Rebind(queryAtt)
	var allAttendances []models.Attendance
	h.DB.Select(&allAttendances, queryAtt, argsAtt...)

	attMap := make(map[int]map[string]models.Attendance)
	for _, att := range allAttendances {
		if _, ok := attMap[att.EmployeeID]; !ok {
			attMap[att.EmployeeID] = make(map[string]models.Attendance)
		}
		dateKey := ""
		if att.CheckIn != nil {
			dateKey = att.CheckIn.Format("2006-01-02")
		} else if att.CreatedAt != nil {
			dateKey = att.CreatedAt.Format("2006-01-02")
		}
		
		if dateKey != "" {
			// Protection: If we already have an attendance for this day, 
			// keep the one that is more "complete" (has check_in/out)
			existing, exists := attMap[att.EmployeeID][dateKey]
			if !exists || (existing.CheckIn == nil && att.CheckIn != nil) || 
			   (existing.CheckOut == nil && att.CheckOut != nil) {
				attMap[att.EmployeeID][dateKey] = att
			}
		}
	}

	// Collect unique attendance IDs for incident lookup
	attIDs := make([]int, 0)
	for _, dayMap := range attMap {
		for _, att := range dayMap {
			attIDs = append(attIDs, att.ID)
		}
	}

	// Carga masiva de incidentes
	incMap := make(map[int][]models.Incident)
	if len(attIDs) > 0 {
		queryInc, argsInc, _ := sqlx.In(`SELECT * FROM incidents WHERE attendance_id IN (?)`, attIDs)
		queryInc = h.DB.Rebind(queryInc)
		var allIncidents []models.Incident
		err = h.DB.Unsafe().Select(&allIncidents, queryInc, argsInc...)
		if err != nil {
			utils.Logger.Error("Error fetching incidents for report job", zap.Int("jobID", jobID), zap.Error(err))
		} else {
			for _, inc := range allIncidents {
				incMap[inc.AttendanceID] = append(incMap[inc.AttendanceID], inc)
			}
		}
	}

	// Procesamiento secuencial con actualización de progreso
	processed := 0
	for _, emp := range employees {
		// Iniciar transacción por empleado para asegurar ACID en el borrado/inserción del reporte
		tx, err := h.DB.Beginx()
		if err != nil {
			utils.Logger.Error("Error starting transaction for employee report", zap.Int("empID", emp.ID), zap.Error(err))
			continue
		}

		empAtts := attMap[emp.ID]
		if empAtts == nil {
			empAtts = make(map[string]models.Attendance)
		}

		var breakdown []models.DailyBreakdownItem
		var totalHours, totalEarnings, totalDeductions float64
		var totalIncidents, daysWorked int

		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			dateStr := d.Format("2006-01-02")
			att, hasAtt := empAtts[dateStr]
			_, isHoliday := holidayMap[dateStr]
			if !hasAtt {
				continue
			}

			item := models.DailyBreakdownItem{
				Date:           dateStr,
				WorkCenterName: emp.WorkCenterName,
				IsHoliday:      isHoliday,
			}
			daysWorked++
			item.CheckIn = formatTime(att.CheckIn)
			item.CheckOut = formatTime(att.CheckOut)
			
			if att.LunchStart != nil && att.LunchEnd != nil {
				item.Lunch = fmt.Sprintf("%s - %s", att.LunchStart.Format("15:04"), att.LunchEnd.Format("15:04"))
			} else {
				item.Lunch = "-"
			}

			if att.CheckIn != nil && att.CheckOut != nil {
				netHValue := 0.0
				if att.NetHoursWorked != nil {
					netHValue = *att.NetHoursWorked
				}
				earnValue := 0.0
				if att.DailyEarnings != nil {
					earnValue = *att.DailyEarnings
				}
				netH := fmt.Sprintf("%.2f", netHValue)
				earn := fmt.Sprintf("%.2f", earnValue)
				item.NetHours = &netH
				item.Earnings = &earn
				totalHours += netHValue
				totalEarnings += earnValue
			} else {
				item.IsIncomplete = true
			}

			dayDeduction := 0.0
			for _, inc := range incMap[att.ID] {
				if inc.Status == models.StatusJustified {
					continue
				}
				totalIncidents++
				if inc.Type == models.IncidentTypeLate {
					dayDeduction += emp.LatePenalty
				} else if inc.Type == models.IncidentTypeOutOfRange {
					dayDeduction += emp.OutOfRangePenalty
				} else if inc.Type == models.IncidentTypeLunchOverstay {
					dayDeduction += emp.LunchPenalty
				}
			}
			item.Deduction = dayDeduction
			totalDeductions += dayDeduction
			breakdown = append(breakdown, item)
		}

		breakdownJSON, _ := json.Marshal(breakdown)
		
		// 1. Eliminar reportes antiguos en el mismo rango (Atomicidad)
		_, err = tx.Exec("DELETE FROM reports WHERE employee_id = $1 AND start_date = $2 AND end_date = $3", emp.ID, start, end)
		if err != nil {
			tx.Rollback()
			utils.Logger.Error("Error deleting old reports for employee", zap.Int("empID", emp.ID), zap.Error(err))
			continue
		}

		// 2. Insertar el nuevo reporte
		_, err = tx.Exec(`
			INSERT INTO reports (
				employee_id, work_center_id, start_date, end_date, 
				total_hours_worked, total_earnings, total_deductions, 
				total_incidents, days_worked, daily_breakdown, status
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'valid')
		`, emp.ID, emp.WorkCenterID, start, end, totalHours, totalEarnings, totalDeductions, totalIncidents, daysWorked, breakdownJSON)

		if err != nil {
			tx.Rollback()
			utils.Logger.Error("Error inserting report for employee", zap.Int("empID", emp.ID), zap.Error(err))
			continue
		}

		// Confirmar transacción por empleado
		if err := tx.Commit(); err != nil {
			utils.Logger.Error("Error committing report for employee", zap.Int("empID", emp.ID), zap.Error(err))
			continue
		}

		processed++
		progress := int((float64(processed) / float64(totalEmps)) * 100)
		if processed%5 == 0 || processed == totalEmps {
			h.DB.Exec("UPDATE report_jobs SET processed_records = $1, progress = $2, updated_at = NOW() WHERE id = $3", processed, progress, jobID)
		}
	}

	// Final generation of files (PDF & Excel)
	h.DB.Exec("UPDATE report_jobs SET progress = 95, status = 'generating_files' WHERE id = $1", jobID)

	var finalReports []models.Report
	err = h.DB.Select(&finalReports, `
		SELECT r.*, u.name as employee_name 
		FROM reports r
		JOIN employees e ON r.employee_id = e.id
		JOIN users u ON e.user_id = u.id
		WHERE r.start_date = $1 AND r.end_date = $2
		ORDER BY u.name ASC
	`, start, end)

	if err == nil && len(finalReports) > 0 {
		dtos := make([]models.ReportDTO, len(finalReports))
		for i, r := range finalReports {
			dtos[i] = models.MapReportToDTO(r)
		}

		var pdfURL, excelURL *string

		// 1. PDF Generation & Upload
		pdfBytes, err := h.PDFService.GenerateReportsPDF(dtos, "vertical", "es")
		if err == nil && h.Storage != nil {
			pdfName := fmt.Sprintf("report_%d_%s_%s.pdf", jobID, req.StartDate, req.EndDate)
			url, err := h.Storage.UploadFile(context.Background(), bytes.NewReader(pdfBytes), pdfName, "application/pdf")
			if err == nil {
				pdfURL = &url
			}
		}

		// 2. Excel Generation & Upload
		excelReader, err := h.ExcelService.GenerateReportsExcel(dtos)
		if err == nil && h.Storage != nil {
			excelName := fmt.Sprintf("report_%d_%s_%s.xlsx", jobID, req.StartDate, req.EndDate)
			url, err := h.Storage.UploadFile(context.Background(), excelReader, excelName, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
			if err == nil {
				excelURL = &url
			}
		}

		h.DB.Exec(`
			UPDATE report_jobs 
			SET status = 'completed', progress = 100, pdf_url = $1, excel_url = $2, updated_at = NOW() 
			WHERE id = $3
		`, pdfURL, excelURL, jobID)
	} else {
		h.DB.Exec("UPDATE report_jobs SET status = 'completed', progress = 100, updated_at = NOW() WHERE id = $1", jobID)
	}

	utils.Logger.Info("Report Job completed successfully", zap.Int("jobID", jobID))
}

func formatTime(t *time.Time) string {
	if t == nil {
		return "N/A"
	}
	return t.Format("15:04")
}

// DeleteReports handles deleting all reports for a specific date range (Transactional)
func (h *ReportHandler) DeleteReports(c *fiber.Ctx) error {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "start_date y end_date son requeridos"})
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al iniciar transacción"})
	}
	defer tx.Rollback()

	result, err := tx.Exec("DELETE FROM reports WHERE start_date = $1 AND end_date = $2", startDate, endDate)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	rows, _ := result.RowsAffected()
	
	userID := c.Locals("user_id").(int)
	err = h.AuditService.LogAction(c.Context(), tx, userID, models.AuditActionDeleteReport, "report", 0, fiber.Map{"start": startDate, "end": endDate}, nil, c.IP())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al registrar auditoría"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al confirmar eliminación"})
	}

	return c.JSON(fiber.Map{"message": fmt.Sprintf("Se eliminaron %d reportes del rango %s - %s", rows, startDate, endDate)})
}

// DeleteReportByID handles deleting a single report by its ID (Transactional)
func (h *ReportHandler) DeleteReportByID(c *fiber.Ctx) error {
	id := c.Params("id")

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al iniciar transacción"})
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM reports WHERE id = $1", id)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	userID := c.Locals("user_id").(int)
	idInt, err := strconv.Atoi(id)
	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{Code: models.ErrInvalidID})
	}
	err = h.AuditService.LogAction(c.Context(), tx, userID, models.AuditActionDeleteReport, "report", idInt, nil, nil, c.IP())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al registrar auditoría"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error al confirmar eliminación"})
	}

	return c.JSON(fiber.Map{"message": "Reporte eliminado exitosamente"})
}




