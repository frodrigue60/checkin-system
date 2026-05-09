package models

// Incident Types
const (
	IncidentTypeLate          = "late"
	IncidentTypeOutOfRange    = "out_of_range"
	IncidentTypeLunchOverstay = "lunch_overstay"
	IncidentTypeAbsent        = "absent"
)

// Incident Status
const (
	StatusPending   = "pending"
	StatusApproved  = "approved"
	StatusJustified = "justified"
)

// Holiday Types
const (
	HolidayTypeMandatory = "mandatory"
	HolidayTypeOptional  = "optional"
)

// Audit Actions
const (
	AuditActionUpdateAttendance     = "UPDATE_ATTENDANCE"
	AuditActionDeleteAttendance     = "DELETE_ATTENDANCE"
	AuditActionUpdateIncidentStatus = "UPDATE_INCIDENT_STATUS"
	
	AuditActionCreateEmployee       = "CREATE_EMPLOYEE"
	AuditActionUpdateEmployee       = "UPDATE_EMPLOYEE"
	AuditActionDeleteEmployee       = "DELETE_EMPLOYEE"
	
	AuditActionCreateWorkCenter     = "CREATE_WORK_CENTER"
	AuditActionUpdateWorkCenter     = "UPDATE_WORK_CENTER"
	AuditActionDeleteWorkCenter     = "DELETE_WORK_CENTER"
	
	AuditActionCreateShift          = "CREATE_SHIFT"
	AuditActionUpdateShift          = "UPDATE_SHIFT"
	AuditActionDeleteShift          = "DELETE_SHIFT"
	
	AuditActionCreatePosition       = "CREATE_POSITION"
	AuditActionUpdatePosition       = "UPDATE_POSITION"
	AuditActionDeletePosition       = "DELETE_POSITION"
	
	AuditActionCreateHoliday        = "CREATE_HOLIDAY"
	AuditActionUpdateHoliday        = "UPDATE_HOLIDAY"
	AuditActionDeleteHoliday        = "DELETE_HOLIDAY"
	
	AuditActionGenerateReport       = "GENERATE_REPORT"
	AuditActionDeleteReport         = "DELETE_REPORT"

	AuditActionBulkUpdateEmployees   = "BULK_UPDATE_EMPLOYEES"
	AuditActionBulkDeleteEmployees   = "BULK_DELETE_EMPLOYEES"
	AuditActionBulkJustifyAttendances = "BULK_JUSTIFY_ATTENDANCES"
	AuditActionCreateUser            = "CREATE_USER"
)

// Role slugs — use these instead of numeric IDs in queries.
// Open/Closed Principle: add new roles without changing existing queries.
//
// Usage in queries:
//   "WHERE role_id = (SELECT id FROM roles WHERE slug = $1)", models.RoleSlugManager
const (
	RoleSlugAdmin      = "admin"
	RoleSlugManager    = "manager"
	RoleSlugSupervisor = "supervisor"
	RoleSlugEmployee   = "employee"
	RoleSlugUser       = "user"
)
