package handlers

import (
	"attendance-api/internal/config"
	"attendance-api/internal/services"

	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
)

// AdminBase contains shared dependencies for all admin sub-handlers.
// Facade Pattern: provides a single initialization point for common resources.
// SOLID (D): handlers depend on this struct, injected from main.go.
type AdminBase struct {
	DB                   *sqlx.DB
	Cfg                  *config.Config
	PDFService           *services.PDFService
	AttendanceService    *services.AttendanceService
	AuditService         *services.AuditService
	ReportService        *services.ReportService
	AlertService         *services.AlertService
	JustificationService *services.JustificationService
	Cache                *cache.Cache
}

// CenterHandler handles CRUD operations for Work Centers.
type CenterHandler struct{ AdminBase }

// ShiftHandler handles CRUD operations for Work Shifts.
type ShiftHandler struct{ AdminBase }

// PositionHandler handles CRUD operations for Positions.
type PositionHandler struct{ AdminBase }

// EmployeeAdminHandler handles CRUD operations for Employees (admin side).
type EmployeeAdminHandler struct{ AdminBase }

// HolidayHandler handles CRUD operations for Holidays.
type HolidayHandler struct{ AdminBase }

// AttendanceAdminHandler handles admin-level attendance management.
type AttendanceAdminHandler struct{ AdminBase }

// IncidentHandler handles incident listing and status updates.
type IncidentHandler struct{ AdminBase }

// DashboardHandler handles dashboard stats and compliance endpoints.
type DashboardHandler struct{ AdminBase }

// ExportHandler handles CSV and PDF export endpoints.
type ExportHandler struct{ AdminBase }

// BulkHandler handles bulk operations on employees, attendances, and incidents.
type BulkHandler struct{ AdminBase }
