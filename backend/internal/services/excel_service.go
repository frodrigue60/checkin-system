package services

import (
	"attendance-api/internal/models"
	"fmt"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"
)

type ExcelService struct{}

func NewExcelService() *ExcelService {
	return &ExcelService{}
}

func (s *ExcelService) GenerateReportsExcel(reports []models.ReportDTO) (io.Reader, error) {
	f := excelize.NewFile()
	defer f.Close()

	summarySheet := "Resumen General"
	f.SetSheetName("Sheet1", summarySheet)

	// Define styles
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"4F81BD"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	detailHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"1F4E78"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	// --- Summary Sheet ---
	headers := []string{
		"ID Empleado", "Nombre Empleado", "Fecha Inicio", "Fecha Fin",
		"Días Trabajados", "Total Horas", "Ganancias Totales", "Deducciones", "Incidentes",
	}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(summarySheet, cell, header)
		f.SetCellStyle(summarySheet, cell, cell, headerStyle)
	}

	for i, r := range reports {
		row := i + 2
		f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), r.EmployeeID)
		f.SetCellValue(summarySheet, fmt.Sprintf("B%d", row), r.EmployeeName)
		f.SetCellValue(summarySheet, fmt.Sprintf("C%d", row), r.StartDate)
		f.SetCellValue(summarySheet, fmt.Sprintf("D%d", row), r.EndDate)
		f.SetCellValue(summarySheet, fmt.Sprintf("E%d", row), r.DaysWorked)
		f.SetCellValue(summarySheet, fmt.Sprintf("F%d", row), r.TotalHoursWorked)
		f.SetCellValue(summarySheet, fmt.Sprintf("G%d", row), r.TotalEarnings)
		f.SetCellValue(summarySheet, fmt.Sprintf("H%d", row), r.TotalDeductions)
		f.SetCellValue(summarySheet, fmt.Sprintf("I%d", row), r.TotalIncidents)

		// --- Detail Sheet for Employee ---
		safeName := s.sanitizeSheetName(fmt.Sprintf("%d-%s", r.EmployeeID, r.EmployeeName))
		f.NewSheet(safeName)

		detailHeaders := []string{"Fecha", "Entrada", "Almuerzo", "Salida", "Horas Netas", "Ganancias", "Deducción", "Centro de Trabajo", "Estado"}
		for j, h := range detailHeaders {
			cell, _ := excelize.CoordinatesToCellName(j+1, 1)
			f.SetCellValue(safeName, cell, h)
			f.SetCellStyle(safeName, cell, cell, detailHeaderStyle)
		}

		for j, d := range r.DailyBreakdown {
			dRow := j + 2
			f.SetCellValue(safeName, fmt.Sprintf("A%d", dRow), d.Date)
			f.SetCellValue(safeName, fmt.Sprintf("B%d", dRow), d.CheckIn)
			f.SetCellValue(safeName, fmt.Sprintf("C%d", dRow), d.Lunch)
			f.SetCellValue(safeName, fmt.Sprintf("D%d", dRow), d.CheckOut)
			
			netHours := ""
			if d.NetHours != nil {
				netHours = *d.NetHours
			}
			f.SetCellValue(safeName, fmt.Sprintf("E%d", dRow), netHours)
			
			earnings := ""
			if d.Earnings != nil {
				earnings = *d.Earnings
			}
			f.SetCellValue(safeName, fmt.Sprintf("F%d", dRow), earnings)
			f.SetCellValue(safeName, fmt.Sprintf("G%d", dRow), d.Deduction)
			f.SetCellValue(safeName, fmt.Sprintf("H%d", dRow), d.WorkCenterName)
			
			status := "Completado"
			if d.IsIncomplete {
				status = "Incompleto"
			}
			if d.IsHoliday {
				status += " (Feriado)"
			}
			f.SetCellValue(safeName, fmt.Sprintf("I%d", dRow), status)
		}

		// Auto-fit detail columns
		for j := range detailHeaders {
			colName, _ := excelize.ColumnNumberToName(j + 1)
			f.SetColWidth(safeName, colName, colName, 18)
		}
	}

	// Auto-fit summary columns
	for i := range headers {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(summarySheet, colName, colName, 20)
	}

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

func (s *ExcelService) sanitizeSheetName(name string) string {
	// Characters not allowed in Excel sheet names: \ / ? * [ ] :
	invalidChars := []string{"\\", "/", "?", "*", "[", "]", ":"}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "_")
	}

	// Limit to 31 characters (Excel limit)
	if len(name) > 31 {
		return name[:31]
	}
	return name
}
