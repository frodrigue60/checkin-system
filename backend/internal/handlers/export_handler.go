package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"attendance-api/internal/models"
	"attendance-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func (h *ExportHandler) ExportAttendancesCSV(c *fiber.Ctx) error {
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

func (h *ExportHandler) ExportAttendancesPDF(c *fiber.Ctx) error {
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


