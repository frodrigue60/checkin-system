package handlers

import (
	"attendance-api/internal/models"
	"attendance-api/internal/services"

	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
)

// AttendanceRich is a joined attendance record with employee, center, and position names.
type AttendanceRich struct {
	models.Attendance
	EmployeeName   string `db:"employee_name"`
	WorkCenterName string `db:"center_name"`
	PositionName   string `db:"position_name"`
	IsLate         bool   `db:"is_late"`
}

// AdminHandler retains only the misc methods that don't fit a specific domain.
// The original 2285-line God Object has been decomposed into:
//   - CenterHandler (center_handler.go)
//   - ShiftHandler (shift_handler.go)
//   - PositionHandler (position_handler.go)
//   - EmployeeAdminHandler (employee_admin_handler.go)
//   - HolidayHandler (holiday_handler.go)
//   - AttendanceAdminHandler (attendance_admin_handler.go)
//   - IncidentHandler (incident_handler.go)
//   - DashboardHandler (dashboard_handler.go)
//   - ExportHandler (export_handler.go)
//   - BulkHandler (bulk_handler.go)
//
// Remaining methods on AdminHandler are in misc_admin_handler.go:
//   ListAuditLogs, ListAlerts, MarkAlertAsRead, ListJustifications, ResolveJustification
type AdminHandler struct {
	DB                   *sqlx.DB
	PDFService           *services.PDFService
	AttendanceService    *services.AttendanceService
	AuditService         *services.AuditService
	ReportService        *services.ReportService
	AlertService         *services.AlertService
	JustificationService *services.JustificationService
	Cache                *cache.Cache
}
