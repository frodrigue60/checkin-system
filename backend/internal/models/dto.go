package models

import "time"

// UserDTO for public/administrative profile presentation
type UserDTO struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	Phone     *string `json:"phone"`
	RoleID    int     `json:"role_id"`
	CreatedAt string  `json:"created_at"`
}

// EmployeeDTO for operational workforce management
type EmployeeDTO struct {
	ID           int  `json:"id"`
	UserID       int  `json:"user_id"`
	PositionID   int  `json:"position_id"`
	WorkCenterID int  `json:"work_center_id"`
	WorkShiftID  int  `json:"work_shift_id"`
	IsActive     bool `json:"is_active"`
}

// AttendanceDTO for tracking logs presentation
type AttendanceDTO struct {
	ID              int        `json:"id"`
	EmployeeID      int        `json:"employee_id"`
	Date            string     `json:"date"`
	CheckIn         *time.Time `json:"check_in"`
	LunchStart      *time.Time `json:"lunch_start"`
	LunchEnd        *time.Time `json:"lunch_end"`
	CheckOut        *time.Time `json:"check_out"`
	NetHoursWorked  float64    `json:"net_hours_worked"`
	DailyEarnings   float64    `json:"daily_earnings"`
	CheckInLatitude   *float64   `json:"check_in_latitude"`
	CheckInLongitude  *float64   `json:"check_in_longitude"`
	CheckOutLatitude  *float64   `json:"check_out_latitude"`
	CheckOutLongitude *float64   `json:"check_out_longitude"`
	IsAbsence       bool       `json:"is_absence"`
	AbsenceReason   *string    `json:"absence_reason"`
	EvidenceURL     *string    `json:"evidence_url"`
	CheckOutNote    *string    `json:"check_out_note"`
	CheckOutAddress *string    `json:"check_out_address"`
	EvidenceURLs    []string   `json:"evidence_urls"`
}

// AttendanceDetailDTO for rich administrative listings
type AttendanceDetailDTO struct {
	ID              int        `json:"id"`
	EmployeeID      int        `json:"employee_id"`
	Date            string     `json:"date"`
	CheckIn         *time.Time `json:"check_in"`
	LunchStart      *time.Time `json:"lunch_start"`
	LunchEnd        *time.Time `json:"lunch_end"`
	CheckOut        *time.Time `json:"check_out"`
	NetHoursWorked  float64    `json:"net_hours_worked"`
	DailyEarnings   float64    `json:"daily_earnings"`
	CheckInLatitude   *float64   `json:"check_in_latitude"`
	CheckInLongitude  *float64   `json:"check_in_longitude"`
	CheckOutLatitude  *float64   `json:"check_out_latitude"`
	CheckOutLongitude *float64   `json:"check_out_longitude"`
	IsAbsence       bool       `json:"is_absence"`
	AbsenceReason   *string    `json:"absence_reason"`
	EmployeeName    string     `json:"employee_name"`
	WorkCenterName  string     `json:"work_center_name"`
	PositionName    string     `json:"position_name"`
	IsLate          bool       `json:"is_late"`
	EvidenceURL     *string    `json:"evidence_url"`
	CheckOutNote    *string    `json:"check_out_note"`
	CheckOutAddress *string    `json:"check_out_address"`
	EvidenceURLs    []string   `json:"evidence_urls"`
}

// ReportDTO for complex payroll and attendance audits
type ReportDTO struct {
	ID                int                  `json:"id"`
	EmployeeID        int                  `json:"employee_id"`
	StartDate         string               `json:"start_date"`
	EndDate           string               `json:"end_date"`
	TotalHoursWorked  float64              `json:"total_hours_worked"`
	TotalEarnings     float64              `json:"total_earnings"`
	TotalDeductions   float64              `json:"total_deductions"`
	TotalIncidents    int                  `json:"total_incidents"`
	DaysWorked        int                  `json:"days_worked"`
	DailyBreakdown    []DailyBreakdownItem `json:"daily_breakdown"`
	EmployeeName      string               `json:"employee_name,omitempty"`
	Status            string               `json:"status"`
}

// WorkCenterDTO for location management
type WorkCenterDTO struct {
	ID                    int     `json:"id"`
	Name                  string  `json:"name"`
	Address               *string `json:"address"`
	Latitude              float64 `json:"latitude"`
	Longitude             float64 `json:"longitude"`
	ToleranceRadiusMeters int     `json:"tolerance_radius"`
}

// PositionDTO for job definitions
type PositionDTO struct {
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	HourlyRate     float64 `json:"base_pay"`
	LatePenaltyFee float64 `json:"late_penalty"`
	OutOfRangeFee  float64 `json:"out_of_range_penalty"`
	LunchOverstayFee float64 `json:"lunch_overstay_penalty"`
	EmployeesCount int     `json:"employees_count"`
}

// WorkShiftDTO for scheduling
type WorkShiftDTO struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	ExpectedCheckIn  string `json:"start_time"`
	ExpectedCheckOut string `json:"end_time"`
	AllowedLunchTime string `json:"lunch_duration_limit"`
	ToleranceTime    string `json:"grace_period"`
	IsNightShift      bool   `json:"is_night_shift"`
	IsActive          bool   `json:"is_active"`
	EnforceLateness   bool   `json:"enforce_lateness"`
	EnforceLunchLimit bool   `json:"enforce_lunch_limit"`
	EnforceGeofence   bool   `json:"enforce_geofence"`
	ShiftType         string `json:"shift_type"`
	WorkDays          []int  `json:"work_days"`
}

// HolidayDTO for calendar presentation
type HolidayDTO struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Date        string `json:"date"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// EmployeeDetailDTO for rich listings with joined data
type EmployeeDetailDTO struct {
	EmployeeDTO
	UserName     string  `json:"user_name"`
	Email        string  `json:"email"`
	Phone        *string `json:"phone"`
	JoinedAt     string  `json:"joined_at"`
	CenterName   string  `json:"center_name"`
	ShiftName    *string `json:"shift_name"`
	PositionName string  `json:"position_name"`
	HourlyRate   float64 `json:"hourly_rate"`
}


// WorkCenterDetailDTO for full aggregation view
type WorkCenterDetailDTO struct {
	Center          WorkCenterDTO         `json:"center"`
	Manager         *UserDTO              `json:"manager"`
	Employees       []EmployeeDetailDTO   `json:"employees"`
	RecentAttendance []AttendanceDetailDTO `json:"recent_attendance"`
}

// PositionDetailDTO for role aggregation view
type PositionDetailDTO struct {
	Position        PositionDTO           `json:"position"`
	Employees       []EmployeeDetailDTO   `json:"employees"`
	RecentAttendance []AttendanceDetailDTO `json:"recent_attendance"`
}

// WorkShiftDetailDTO for shift aggregation view
type WorkShiftDetailDTO struct {
	Shift            WorkShiftDTO          `json:"shift"`
	Employees        []EmployeeDetailDTO   `json:"employees"`
	RecentAttendance []AttendanceDetailDTO `json:"recent_attendance"`
}

// EmployeeFullDetailDTO for the main individual profile view
type EmployeeFullDetailDTO struct {
	Employee         EmployeeDetailDTO     `json:"employee"`
	Stats            EmployeeStats         `json:"stats"`
	RecentAttendance []AttendanceDetailDTO `json:"recent_attendance"`
}

// EmployeeStats for aggregate performance metrics
type EmployeeStats struct {
	TotalAttendances int     `json:"total_attendances"`
	TotalHours       float64 `json:"total_hours"`
	TotalEarnings    float64 `json:"total_earnings"`
	TotalIncidents   int     `json:"total_incidents"`
}

// PaginatedResponse wrapper for collections
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

// IncidentDTO for rich infraction logs
type IncidentDTO struct {
	ID           int       `json:"id"`
	Type         string    `json:"type"`
	IsLate       bool      `json:"is_late"`
	DelayMinutes int       `json:"delay_minutes"`
	IsOutOfRange bool      `json:"is_out_of_range"`
	Distance     int       `json:"distance"`
	Status       string    `json:"status"`
	ResolvedBy   *int      `json:"resolved_by"`
	ResolutionNote *string `json:"resolution_note"`
	MetadataJSON *string   `json:"metadata"`
	CreatedAt    *time.Time `json:"created_at"`
}

// AttendanceFullDetailDTO for the individual tracking audit view
type AttendanceFullDetailDTO struct {
	Attendance AttendanceDetailDTO `json:"attendance"`
	WorkCenter WorkCenterDTO       `json:"work_center"`
	Shift      *WorkShiftDTO       `json:"shift"`
	Incidents  []IncidentDTO       `json:"incidents"`
}

// BulkRequest for generic multiple ID actions
type BulkRequest struct {
	IDs []int `json:"ids" validate:"required,min=1"`
}

// BulkEmployeeUpdate for mass attribute changes
type BulkEmployeeUpdate struct {
	IDs          []int `json:"ids" validate:"required,min=1"`
	WorkCenterID *int  `json:"work_center_id,omitempty"`
	WorkShiftID  *int  `json:"work_shift_id,omitempty"`
	IsActive     *bool `json:"is_active,omitempty"`
}

// BulkJustifyRequest for mass incident resolution
type BulkJustifyRequest struct {
	AttendanceIDs []int  `json:"attendance_ids" validate:"required,min=1"`
	Note          string `json:"note" validate:"required"`
}

// BulkJustifyIncidentsRequest for mass incident resolution specifically by incident IDs
type BulkJustifyIncidentsRequest struct {
	IncidentIDs []int  `json:"incident_ids" validate:"required,min=1"`
	Note        string `json:"note" validate:"required"`
}

// IncidentRichDTO for the dedicated incidents view
type IncidentRichDTO struct {
	IncidentDTO
	EmployeeName   string `json:"employee_name"`
	AttendanceDate string `json:"attendance_date"`
	AttendanceID   int    `json:"attendance_id"`
	WorkCenterName string `json:"center_name"`
}
