package models

// API Error Codes — frontend maps these to localized messages via i18n.
// The backend sends stable codes; the frontend is responsible for translation.

// Error code constants
const (
	ErrInvalidID         = "INVALID_ID"
	ErrNotFound          = "NOT_FOUND"
	ErrUnauthorized      = "UNAUTHORIZED"
	ErrForbidden         = "FORBIDDEN"
	ErrInvalidBody       = "INVALID_BODY"
	ErrHolidayBlocked    = "HOLIDAY_BLOCKED"
	ErrSessionActive     = "SESSION_ACTIVE"
	ErrNoActiveSession   = "NO_ACTIVE_SESSION"
	ErrIncompleteProfile = "INCOMPLETE_PROFILE"
	ErrInternalServer    = "INTERNAL_ERROR"
	ErrDuplicateEmail    = "DUPLICATE_EMAIL"
	ErrEmployeeInactive  = "EMPLOYEE_INACTIVE"
	ErrEvidenceRequired  = "EVIDENCE_REQUIRED"
	ErrAlreadyExists     = "ALREADY_EXISTS"
	ErrTransactionFailed = "TRANSACTION_FAILED"
	ErrRegistrationClosed = "REGISTRATION_CLOSED"
)

// APIError is the standardized error response for all API endpoints.
// Frontend uses the Code field for i18n lookup, Message is a fallback.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}
