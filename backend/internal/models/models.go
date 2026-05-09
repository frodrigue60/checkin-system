package models

import (
	"encoding/json"
	"time"
)

type AttendanceExportDTO struct {
	ID             int     `json:"id"`
	EmployeeName   string  `json:"employee_name"`
	CenterName     string  `json:"center_name"`
	PositionName   string  `json:"position_name"`
	CheckIn        string  `json:"check_in"`
	CheckOut       string  `json:"check_out"`
	Hours          float64 `json:"hours"`
	Earnings       float64 `json:"earnings"`
	IsLate         bool    `json:"is_late"`
	IsAbsence      bool    `json:"is_absence"`
	AbsenceReason  string  `json:"absence_reason"`
}

type User struct {
	ID              int        `db:"id" json:"id"`
	Name            string     `db:"name" json:"name"`
	Email           string     `db:"email" json:"email"`
	Phone           *string    `db:"phone" json:"phone"`
	EmailVerifiedAt *time.Time `db:"email_verified_at" json:"email_verified_at"`
	Password        string     `db:"password" json:"-"`
	RememberToken   *string    `db:"remember_token" json:"remember_token"`
	RoleID          int        `db:"role_id" json:"role_id"`
	CreatedAt       *time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       *time.Time `db:"updated_at" json:"updated_at"`
}

type Role struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
	Slug string `db:"slug" json:"slug"`
}

type Employee struct {
	ID           int        `db:"id" json:"id"`
	UserID       int        `db:"user_id" json:"user_id"`
	PositionID   int        `db:"position_id" json:"position_id"`
	WorkCenterID int        `db:"work_center_id" json:"work_center_id"`
	WorkShiftID  int        `db:"work_shift_id" json:"work_shift_id"`
	IsActive     bool       `db:"is_active" json:"is_active"`
	CreatedAt    *time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    *time.Time `db:"updated_at" json:"updated_at"`
}

type WorkCenter struct {
	ID                    int        `db:"id" json:"id"`
	Name                  string     `db:"name" json:"name"`
	Address               *string    `db:"address" json:"address"`
	Latitude              float64    `db:"latitude" json:"latitude"`
	Longitude             float64    `db:"longitude" json:"longitude"`
	ToleranceRadiusMeters int        `db:"tolerance_radius_meters" json:"tolerance_radius"`
	ManagerID             *int       `db:"manager_id" json:"manager_id"`
	Timezone              string     `db:"timezone" json:"timezone"`
	CreatedAt             *time.Time `db:"created_at" json:"created_at"`
	UpdatedAt             *time.Time `db:"updated_at" json:"updated_at"`
}

type WorkShift struct {
	ID                 int        `db:"id" json:"id"`
	Name               string          `db:"name" json:"name" validate:"required"`
	ExpectedCheckIn    string          `db:"expected_check_in" json:"start_time" validate:"required"`
	ExpectedCheckOut   string          `db:"expected_check_out" json:"end_time" validate:"required"`
	AllowedLunchTime   string          `db:"allowed_lunch_time" json:"lunch_duration_limit" validate:"required"`
	ToleranceTime      string          `db:"tolerance_time" json:"grace_period" validate:"required"`
	IsNightShift       bool            `db:"is_night_shift" json:"is_night_shift"`
	IsActive           bool            `db:"is_active" json:"is_active"`
	EnforceLateness    bool            `db:"enforce_lateness" json:"enforce_lateness"`
	EnforceLunchLimit  bool            `db:"enforce_lunch_limit" json:"enforce_lunch_limit"`
	EnforceGeofence    bool            `db:"enforce_geofence" json:"enforce_geofence"`
	ShiftType          string          `db:"shift_type" json:"shift_type" validate:"required,oneof=fixed flexible field"`
	WorkDays           json.RawMessage `db:"work_days" json:"work_days" validate:"required"`
	CreatedAt          *time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt          *time.Time `db:"updated_at" json:"updated_at"`
}

type Position struct {
	ID                int        `db:"id" json:"id"`
	Name              string     `db:"name" json:"name"`
	HourlyRate        float64    `db:"hourly_rate" json:"base_pay"`
	LatePenaltyFee    float64    `db:"late_penalty_fee" json:"late_penalty"`
	OutOfRangeFee     float64    `db:"out_of_range_fee" json:"out_of_range_penalty"`
	LunchOverstayFee  float64    `db:"lunch_overstay_fee" json:"lunch_overstay_penalty"`
	EmployeesCount    int        `db:"employees_count" json:"employees_count"`
	CreatedAt         *time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         *time.Time `db:"updated_at" json:"updated_at"`
}

type Attendance struct {
	ID                int        `db:"id" json:"id"`
	EmployeeID        int        `db:"employee_id" json:"employee_id"`
	WorkShiftID       *int       `db:"work_shift_id" json:"work_shift_id"`
	WorkCenterID      *int       `db:"work_center_id" json:"work_center_id"`
	CheckIn           *time.Time `db:"check_in" json:"check_in"`
	LunchStart        *time.Time `db:"lunch_start" json:"lunch_start"`
	LunchEnd          *time.Time `db:"lunch_end" json:"lunch_end"`
	CheckOut          *time.Time `db:"check_out" json:"check_out"`
	CheckInLatitude   *float64   `db:"check_in_latitude" json:"check_in_latitude"`
	CheckInLongitude  *float64   `db:"check_in_longitude" json:"check_in_longitude"`
	CheckOutLatitude  *float64   `db:"check_out_latitude" json:"check_out_latitude"`
	CheckOutLongitude *float64   `db:"check_out_longitude" json:"check_out_longitude"`
	NetHoursWorked    *float64   `db:"net_hours_worked" json:"net_hours_worked"`
	DailyEarnings     *float64   `db:"daily_earnings" json:"daily_earnings"`
	IsAbsence         bool       `db:"is_absence" json:"is_absence"`
	AbsenceReason     *string    `db:"absence_reason" json:"absence_reason"`
	EvidenceURL       *string    `db:"evidence_url" json:"evidence_url"`
	CheckOutNote      *string    `db:"check_out_note" json:"check_out_note"`
	CheckInAddress    *string    `db:"check_in_address" json:"check_in_address"`
	CheckInNote       *string    `db:"check_in_note" json:"check_in_note"`
	CheckOutAddress   *string         `db:"check_out_address" json:"check_out_address"`
	IsFieldWork       bool            `db:"is_field_work" json:"is_field_work"`
	EvidenceURLs      json.RawMessage `db:"evidence_urls" json:"evidence_urls"`
	CreatedAt         *time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt         *time.Time      `db:"updated_at" json:"updated_at"`
}

type Holiday struct {
	ID          int        `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Date        time.Time  `db:"date" json:"date"`
	Description *string    `db:"description" json:"description"`
	Type        string     `db:"type" json:"type"` // 'mandatory', 'optional'
	CreatedAt   *time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at" json:"updated_at"`
}

type Incident struct {
	ID             int        `db:"id" json:"id"`
	EmployeeID     int        `db:"employee_id" json:"employee_id"`
	WorkCenterID   int        `db:"work_center_id" json:"work_center_id"`
	AttendanceID   int        `db:"attendance_id" json:"attendance_id"`
	Type           string     `db:"type" json:"type"` // late, out_of_range
	Description    *string    `db:"description" json:"description"`
	IsLate         bool       `db:"is_late" json:"is_late"`
	DelayMinutes   int        `db:"delay_minutes" json:"delay_minutes"`
	IsOutOfRange   bool       `db:"is_out_of_range" json:"is_out_of_range"`
	Distance       int        `db:"distance" json:"distance"`
	CheckInTime    *string    `db:"check_in_time" json:"check_in_time"`
	Status         string     `db:"status" json:"status"` // pending, approved, justified
	ResolvedBy     *int       `db:"resolved_by" json:"resolved_by"`
	ResolutionNote *string    `db:"resolution_note" json:"resolution_note"`
	MetadataJSON   *string    `db:"metadata_json" json:"metadata_json"`
	CreatedAt      *time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      *time.Time `db:"updated_at" json:"updated_at"`

	// Relational
	Justification *Justification `db:"-" json:"justification,omitempty"`
}

type Justification struct {
	ID             int        `db:"id" json:"id"`
	IncidentID     int        `db:"incident_id" json:"incident_id"`
	EmployeeID     int        `db:"employee_id" json:"employee_id"`
	Message        string     `db:"message" json:"message"`
	EvidenceURL    *string    `db:"evidence_url" json:"evidence_url"`
	Status         string     `db:"status" json:"status"` // pending, approved, rejected
	ResolvedBy     *int       `db:"resolved_by" json:"resolved_by"`
	ResolutionNote *string    `db:"resolution_note" json:"resolution_note"`
	CreatedAt      *time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      *time.Time `db:"updated_at" json:"updated_at"`
}

type DailyBreakdownItem struct {
	Date           string  `json:"date"`
	CheckIn        string  `json:"check_in"`
	Lunch          string  `json:"lunch"`
	CheckOut       string  `json:"check_out"`
	NetHours       *string `json:"net_hours"`
	Earnings       *string `json:"earnings"`
	Deduction      float64 `json:"deduction"`
	WorkCenterName string  `json:"work_center_name"`
	IsIncomplete   bool    `json:"is_incomplete"`
	IsHoliday      bool    `json:"is_holiday"`
}

type Report struct {
	ID                int       `db:"id" json:"-"`
	EmployeeID        int       `db:"employee_id" json:"-"`
	WorkCenterID      *int      `db:"work_center_id" json:"-"`
	StartDate         time.Time `db:"start_date" json:"-"`
	EndDate           time.Time `db:"end_date" json:"-"`
	TotalHoursWorked  float64   `db:"total_hours_worked" json:"-"`
	TotalEarnings     float64   `db:"total_earnings" json:"-"`
	TotalDeductions   float64   `db:"total_deductions" json:"-"`
	TotalIncidents    int       `db:"total_incidents" json:"-"`
	DaysWorked        int       `db:"days_worked" json:"-"`
	DailyBreakdown    []byte     `db:"daily_breakdown" json:"-"`
	EmployeeName      string     `db:"employee_name" json:"-"`
	Status            string     `db:"status" json:"-"`
	CreatedAt         *time.Time `db:"created_at" json:"-"`
	UpdatedAt         *time.Time `db:"updated_at" json:"-"`
}

// AuditLog represents a record in the audit_logs table
type AuditLog struct {
	ID         int             `db:"id" json:"id"`
	UserID     *int            `db:"user_id" json:"user_id"`
	Action     string          `db:"action" json:"action"`
	EntityType string          `db:"entity_type" json:"entity_type"`
	EntityID   *int            `db:"entity_id" json:"entity_id"`
	OldValue   *string         `db:"old_value" json:"old_value"` // JSON string
	NewValue   *string         `db:"new_value" json:"new_value"` // JSON string
	IPAddress  *string         `db:"ip_address" json:"ip_address"`
	CreatedAt  time.Time       `db:"created_at" json:"created_at"`
	
	// Virtual fields for display
	UserName   string `db:"user_name" json:"user_name,omitempty"`
}

type SystemAlert struct {
	ID           int             `db:"id" json:"id"`
	Type         string          `db:"type" json:"type"`
	Severity     string          `db:"severity" json:"severity"`
	Message      string          `db:"message" json:"message"`
	MetadataJSON *string         `db:"metadata_json" json:"metadata_json"`
	IsRead       bool            `db:"is_read" json:"is_read"`
	CreatedAt    time.Time       `db:"created_at" json:"created_at"`
}

type ReportJob struct {
	ID               int       `db:"id" json:"id"`
	UserID           int       `db:"user_id" json:"user_id"`
	StartDate        time.Time `db:"start_date" json:"start_date"`
	EndDate          time.Time `db:"end_date" json:"end_date"`
	Status           string    `db:"status" json:"status"`
	Progress         int       `db:"progress" json:"progress"`
	TotalRecords     int       `db:"total_records" json:"total_records"`
	ProcessedRecords int       `db:"processed_records" json:"processed_records"`
	ErrorMessage     *string   `db:"error_message" json:"error_message"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

type ReportJobDTO struct {
	ID               int    `json:"id"`
	Status           string `json:"status"`
	Progress         int    `json:"progress"`
	ProcessedRecords int    `json:"processed_records"`
	TotalRecords     int    `json:"total_records"`
	StartDate        string `json:"start_date"`
	EndDate          string `json:"end_date"`
	CreatedAt        string `json:"created_at"`
}
