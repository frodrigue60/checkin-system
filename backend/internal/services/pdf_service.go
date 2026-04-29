package services

import (
	"attendance-api/internal/models"
	"bytes"
	"fmt"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/johnfercher/maroto/v2/pkg/components/page"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
)

type PDFService struct {
	appName string
}

func NewPDFService(appName string) *PDFService {
	return &PDFService{appName: appName}
}

type PDFLabels struct {
	Title          string
	Subtitle       string
	Employee       string
	Period         string
	Consolidated   string
	DaysWorked     string
	TotalHours     string
	Deductions     string
	NetEarnings    string
	Breakdown      string
	Date           string
	CheckIn        string
	CheckOut       string
	Hours          string
	Earnings       string
	DeductionShort string
	NoRecords      string
}

func getLabels(lang string) PDFLabels {
	if lang == "en" {
		return PDFLabels{
			Title:          "Attendance & Payroll Report",
			Subtitle:       "Operational Performance Document",
			Employee:       "Employee",
			Period:         "Period",
			Consolidated:   "Consolidated Summary",
			DaysWorked:     "Days Worked",
			TotalHours:     "Total Hours",
			Deductions:     "Deductions",
			NetEarnings:    "Net Earnings",
			Breakdown:      "Daily Breakdown",
			Date:           "Date",
			CheckIn:        "Check-in",
			CheckOut:       "Check-out",
			Hours:          "Hours",
			Earnings:       "Earnings",
			DeductionShort: "Deduct.",
			NoRecords:      "No valid attendance records found for the selected period.",
		}
	}
	// Default Spanish
	return PDFLabels{
		Title:          "Reporte de Asistencia y Nómina",
		Subtitle:       "Documento de Desempeño Operativo",
		Employee:       "Empleado",
		Period:         "Período",
		Consolidated:   "Resumen Consolidado",
		DaysWorked:     "Días Trabajados",
		TotalHours:     "Horas Totales",
		Deductions:     "Deducciones",
		NetEarnings:    "Ganancia Neta",
		Breakdown:      "Desglose Diario",
		Date:           "Fecha",
		CheckIn:        "Entrada",
		CheckOut:       "Salida",
		Hours:          "Horas",
		Earnings:       "Ganancia",
		DeductionShort: "Deducc.",
		NoRecords:      "No hay registros de asistencia válidos para el periodo seleccionado.",
	}
}

func (s *PDFService) GenerateReportsPDF(reports []models.ReportDTO, orientationStr string, lang string) ([]byte, error) {
	cfgBuilder := config.NewBuilder()

	if orientationStr == "horizontal" {
		cfgBuilder.WithOrientation(orientation.Horizontal)
	} else {
		cfgBuilder.WithOrientation(orientation.Vertical)
	}

	m := maroto.New(cfgBuilder.Build())
	labels := getLabels(lang)

	for _, report := range reports {
		rows := s.getReportRows(report, labels)
		m.AddPages(page.New().Add(rows...))
	}

	document, err := m.Generate()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.Write(document.GetBytes())
	return buf.Bytes(), nil
}

func (s *PDFService) getReportRows(report models.ReportDTO, labels PDFLabels) []core.Row {
	var rows []core.Row

	// Header Section
	rows = append(rows, row.New(20).Add(
		col.New(12).Add(
			text.New(s.appName, props.Text{
				Size:  20,
				Style: fontstyle.Bold,
				Align: align.Center,
			}),
		),
	))

	rows = append(rows, row.New(10).Add(
		col.New(12).Add(
			text.New(labels.Title, props.Text{
				Size:  12,
				Style: fontstyle.Italic,
				Align: align.Center,
			}),
		),
	))

	rows = append(rows, row.New(10).Add(col.New(12))) // Space

	// Employee Info
	rows = append(rows, row.New(10).Add(
		col.New(6).Add(
			text.New(fmt.Sprintf("%s: %s", labels.Employee, report.EmployeeName), props.Text{Style: fontstyle.Bold}),
		),
		col.New(6).Add(
			text.New(fmt.Sprintf("%s: %s - %s", labels.Period, report.StartDate, report.EndDate), props.Text{Align: align.Right}),
		),
	))

	rows = append(rows, row.New(10).Add(col.New(12))) // Space

	// Financial Summary Table
	rows = append(rows, row.New(10).Add(
		col.New(12).Add(text.New(labels.Consolidated, props.Text{Size: 10, Style: fontstyle.Bold})),
	))

	rows = append(rows, row.New(12).Add(
		col.New(3).Add(text.New(labels.DaysWorked, props.Text{Size: 9, Style: fontstyle.Bold})),
		col.New(3).Add(text.New(labels.TotalHours, props.Text{Size: 9, Style: fontstyle.Bold})),
		col.New(3).Add(text.New(labels.Deductions, props.Text{Size: 9, Style: fontstyle.Bold})),
		col.New(3).Add(text.New(labels.NetEarnings, props.Text{Size: 9, Style: fontstyle.Bold})),
	))

	rows = append(rows, row.New(10).Add(
		col.New(3).Add(text.New(fmt.Sprintf("%d", report.DaysWorked), props.Text{Size: 9})),
		col.New(3).Add(text.New(fmt.Sprintf("%.2f", report.TotalHoursWorked), props.Text{Size: 9})),
		col.New(3).Add(text.New(fmt.Sprintf("$%.2f", report.TotalDeductions), props.Text{Size: 9})),
		col.New(3).Add(text.New(fmt.Sprintf("$%.2f", report.TotalEarnings), props.Text{Size: 9})),
	))

	rows = append(rows, row.New(15).Add(col.New(12))) // Space

	// Detailed Breakdown
	if len(report.DailyBreakdown) > 0 {
		rows = append(rows, row.New(10).Add(
			col.New(12).Add(text.New(labels.Breakdown, props.Text{Size: 10, Style: fontstyle.Bold})),
		))

		// Table Header
		rows = append(rows, row.New(10).Add(
			col.New(2).Add(text.New(labels.Date, props.Text{Size: 8, Style: fontstyle.Bold})),
			col.New(2).Add(text.New(labels.CheckIn, props.Text{Size: 8, Style: fontstyle.Bold})),
			col.New(2).Add(text.New(labels.CheckOut, props.Text{Size: 8, Style: fontstyle.Bold})),
			col.New(2).Add(text.New(labels.Hours, props.Text{Size: 8, Style: fontstyle.Bold})),
			col.New(2).Add(text.New(labels.Earnings, props.Text{Size: 8, Style: fontstyle.Bold})),
			col.New(2).Add(text.New(labels.DeductionShort, props.Text{Size: 8, Style: fontstyle.Bold})),
		))

		for _, item := range report.DailyBreakdown {
			netHours := "0.00"
			if item.NetHours != nil {
				netHours = *item.NetHours
			}
			earnings := "0.00"
			if item.Earnings != nil {
				earnings = *item.Earnings
			}

			rows = append(rows, row.New(8).Add(
				col.New(2).Add(text.New(item.Date, props.Text{Size: 8})),
				col.New(2).Add(text.New(item.CheckIn, props.Text{Size: 8})),
				col.New(2).Add(text.New(item.CheckOut, props.Text{Size: 8})),
				col.New(2).Add(text.New(netHours, props.Text{Size: 8})),
				col.New(2).Add(text.New(fmt.Sprintf("$%s", earnings), props.Text{Size: 8})),
				col.New(2).Add(text.New(fmt.Sprintf("$%.2f", item.Deduction), props.Text{Size: 8})),
			))
		}
	} else {
		rows = append(rows, row.New(20).Add(
			col.New(12).Add(text.New(labels.NoRecords, props.Text{
				Size:  10,
				Style: fontstyle.Italic,
				Align: align.Center,
			})),
		))
	}

	return rows
}

func (s *PDFService) GenerateAttendanceLogPDF(attendances []models.AttendanceExportDTO, lang string) ([]byte, error) {
	m := maroto.New()
	labels := getLabels(lang)

	m.AddRows(row.New(20).Add(
		col.New(12).Add(
			text.New(labels.Title, props.Text{
				Size:  16,
				Style: fontstyle.Bold,
				Align: align.Center,
			}),
		),
	))

	m.AddRows(row.New(10).Add(
		col.New(12).Add(
			text.New(fmt.Sprintf("Generado el: %s", time.Now().Format("2006-01-02 15:04")), props.Text{
				Size:  8,
				Style: fontstyle.Italic,
				Align: align.Right,
			}),
		),
	))

	// Table Header
	m.AddRows(row.New(12).Add(
		col.New(3).Add(text.New("Empleado", props.Text{Size: 9, Style: fontstyle.Bold})),
		col.New(2).Add(text.New("Sede", props.Text{Size: 9, Style: fontstyle.Bold})),
		col.New(2).Add(text.New("Puesto", props.Text{Size: 9, Style: fontstyle.Bold})),
		col.New(2).Add(text.New("Entrada/Fecha", props.Text{Size: 9, Style: fontstyle.Bold})),
		col.New(1).Add(text.New("Salida", props.Text{Size: 9, Style: fontstyle.Bold})),
		col.New(2).Add(text.New("Estado", props.Text{Size: 9, Style: fontstyle.Bold})),
	))

	for _, a := range attendances {
		status := "OK"
		if a.IsLate {
			status = "Tarde"
		}
		if a.IsAbsence {
			status = "Ausencia"
		}

		m.AddRows(row.New(10).Add(
			col.New(3).Add(text.New(a.EmployeeName, props.Text{Size: 8})),
			col.New(2).Add(text.New(a.CenterName, props.Text{Size: 8})),
			col.New(2).Add(text.New(a.PositionName, props.Text{Size: 8})),
			col.New(2).Add(text.New(a.CheckIn, props.Text{Size: 8})),
			col.New(1).Add(text.New(a.CheckOut, props.Text{Size: 8})),
			col.New(2).Add(text.New(status, props.Text{Size: 8})),
		))
	}

	doc, err := m.Generate()
	if err != nil {
		return nil, err
	}

	return doc.GetBytes(), nil
}
