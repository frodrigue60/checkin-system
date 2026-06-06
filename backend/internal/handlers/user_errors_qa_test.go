package handlers

import (
	"attendance-api/internal/config"
	"attendance-api/internal/models"
	"attendance-api/internal/services"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
)

func setupTestErrorHandlers(t *testing.T) (*fiber.App, sqlmock.Sqlmock, *ShiftHandler, *EmployeeAdminHandler) {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if len(c.Response().Body()) > 0 {
				return nil
			}
			return fiber.DefaultErrorHandler(c, err)
		},
	})

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", 100)
		return c.Next()
	})

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening mock db: %s", err)
	}
	sqlxDB := sqlx.NewDb(db, "postgres")

	cCache := cache.New(5*time.Minute, 10*time.Minute)
	cfg := &config.Config{
		JWTSecret:   "test-secret",
		R2PublicURL: "http://localhost/r2",
	}

	auditService := &services.AuditService{DB: sqlxDB}
	base := AdminBase{
		DB:           sqlxDB,
		Cfg:          cfg,
		AuditService: auditService,
		Cache:        cCache,
	}

	shiftHandler := &ShiftHandler{AdminBase: base}
	empHandler := &EmployeeAdminHandler{AdminBase: base}

	return app, mock, shiftHandler, empHandler
}

// ==========================================
// 1. Malformed JSON Body (User sends garbage)
// ==========================================
func TestQAUserErrors_MalformedJSON(t *testing.T) {
	app, mock, shiftHandler, _ := setupTestErrorHandlers(t)
	app.Post("/admin/shifts", shiftHandler.CreateShift)

	// User sends raw broken string instead of valid JSON
	req := httptest.NewRequest("POST", "/admin/shifts", bytes.NewBufferString("{broken-json: true,"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)

	// Fiber's BodyParser or validation should return 400 Bad Request
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==========================================
// 2. Missing required fields (User sends empty shift fields)
// ==========================================
func TestQAUserErrors_MissingRequiredFields(t *testing.T) {
	app, mock, shiftHandler, _ := setupTestErrorHandlers(t)
	app.Post("/admin/shifts", shiftHandler.CreateShift)

	// User leaves fields blank
	invalidShift := models.WorkShift{
		Name:            "", // Required
		ExpectedCheckIn: "", // Required
	}
	body, _ := json.Marshal(invalidShift)

	req := httptest.NewRequest("POST", "/admin/shifts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)

	// Should fail validator tags with 400 Bad Request
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	
	// Check validation error response format
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Contains(t, result, "errors")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==========================================
// 3. Invalid Enum Value (User sends weird ShiftType)
// ==========================================
func TestQAUserErrors_InvalidShiftTypeEnum(t *testing.T) {
	app, mock, shiftHandler, _ := setupTestErrorHandlers(t)
	app.Post("/admin/shifts", shiftHandler.CreateShift)

	daysRaw, _ := json.Marshal([]int{1, 2, 3})
	invalidShift := models.WorkShift{
		Name:             "Morning Shift",
		ExpectedCheckIn:  "09:00",
		ExpectedCheckOut: "18:00",
		AllowedLunchTime: "01:00",
		ToleranceTime:    "00:15",
		ShiftType:        "extraneous_type_that_is_not_oneof", // Must be fixed, flexible, or field
		WorkDays:         daysRaw,
	}
	body, _ := json.Marshal(invalidShift)

	req := httptest.NewRequest("POST", "/admin/shifts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)

	// Should fail with 400 Bad Request due to validation on ShiftType oneof
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==========================================
// 4. Duplicate Employee (DB Unique constraint error)
// ==========================================
func TestQAUserErrors_DuplicateEmployeeID(t *testing.T) {
	app, mock, _, empHandler := setupTestErrorHandlers(t)
	app.Post("/admin/employees", empHandler.CreateEmployee)

	newEmp := models.Employee{
		UserID:       10, // already employed!
		PositionID:   1,
		WorkCenterID: 2,
		WorkShiftID:  3,
	}
	body, _ := json.Marshal(newEmp)

	mock.ExpectBegin()
	// Insert query fails with unique constraint violation (database unique error)
	mock.ExpectQuery("INSERT INTO employees").
		WillReturnError(fmt.Errorf("pq: duplicate key value violates unique constraint \"employees_user_id_key\""))
	mock.ExpectRollback()

	req := httptest.NewRequest("POST", "/admin/employees", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)

	// Handlers must return 500 Internal Server Error when database queries fail
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==========================================
// 5. Nonexistent WorkShift (DB Foreign Key error)
// ==========================================
func TestQAUserErrors_NonexistentWorkShiftAssign(t *testing.T) {
	app, mock, _, empHandler := setupTestErrorHandlers(t)
	app.Put("/admin/employees/:id", empHandler.UpdateEmployee)

	updateEmp := models.Employee{
		WorkCenterID: 2,
		WorkShiftID:  9999, // Nonexistent shift ID!
		PositionID:   1,
		IsActive:     true,
	}
	body, _ := json.Marshal(updateEmp)

	mock.ExpectBegin()
	// Get old state
	mock.ExpectQuery("SELECT \\* FROM employees WHERE id = \\$1").
		WithArgs(40).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "work_center_id", "work_shift_id", "position_id", "is_active"}).
			AddRow(40, 10, 2, 3, 1, true))

	// Executing update query triggers foreign key violation on work_shift_id
	mock.ExpectExec("UPDATE employees SET").
		WillReturnError(fmt.Errorf("pq: insert or update on table \"employees\" violates foreign key constraint \"employees_work_shift_id_fkey\""))
	mock.ExpectRollback()

	req := httptest.NewRequest("PUT", "/admin/employees/40", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)

	// Handlers must return 500 Internal Server Error when DB constraints are violated on save
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==========================================
// 6. Archaic Shift Times & Format (User inputs weird dates/times)
// ==========================================
func TestQAUserErrors_ArchaicShiftTimeFormats(t *testing.T) {
	app, mock, shiftHandler, _ := setupTestErrorHandlers(t)
	app.Post("/admin/shifts", shiftHandler.CreateShift)

	// User enters invalid strings representing times or legacy year-0 placeholder
	daysRaw, _ := json.Marshal([]int{1, 2, 3})
	invalidShift := models.WorkShift{
		Name:             "Night Shift",
		ExpectedCheckIn:  "0001-01-01T00:00:00Z", // Archaic timestamp
		ExpectedCheckOut: "0001-01-01T08:00:00Z", // Archaic timestamp
		AllowedLunchTime: "lunch time",          // Completely malformed
		ToleranceTime:    "none",                // Completely malformed
		ShiftType:        "fixed",
		WorkDays:         daysRaw,
	}
	body, _ := json.Marshal(invalidShift)

	mock.ExpectBegin()
	// CreateShift named query expects insertion
	mock.ExpectQuery("INSERT INTO work_shifts").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(20))
	mock.ExpectExec("INSERT INTO audit_logs").
		WithArgs(100, string(models.AuditActionCreateShift), "work_shift", 20, nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req := httptest.NewRequest("POST", "/admin/shifts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)

	// In current implementation, there's no format check at input validation, so it succeeds in creating it.
	// But our service handles it safely during operation (as tested in attendance_qa_test.go).
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==========================================
// 7. Check-out without required evidence URL (User forgets evidence)
// ==========================================
func TestQAUserErrors_CheckOutMissingEvidence(t *testing.T) {
	app, mock, shiftHandler, _ := setupTestErrorHandlers(t)

	// Setup AttendanceHandler using the same database mock
	attHandler := &AttendanceHandler{
		DB:      shiftHandler.DB,
		Cfg:     shiftHandler.Cfg,
		Service: &services.AttendanceService{},
	}
	app.Post("/attendance/checkout", attHandler.CheckOutNoID)

	// User sends checkout request but evidence_urls is empty (requires at least 2)
	checkoutReq := AttendanceRequest{
		EmployeeID:   10,
		Latitude:     19.4326,
		Longitude:    -99.1332,
		EvidenceURLs: []string{}, // Forgot to upload photo evidence
	}
	body, _ := json.Marshal(checkoutReq)

	// Mock employee and attendance queries since checkout handler runs them first
	mock.ExpectQuery("SELECT \\* FROM employees WHERE user_id = \\$1").
		WithArgs(100). // user_id is set to 100 in setupTestErrorHandlers
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "is_active"}).AddRow(10, 100, true))

	mock.ExpectQuery("SELECT \\* FROM attendances WHERE employee_id = \\$1 AND check_out IS NULL").
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "employee_id", "check_in"}).AddRow(50, 10, time.Now()))

	req := httptest.NewRequest("POST", "/attendance/checkout", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)

	// Should return 400 Bad Request
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Contains(t, result["error"], "evidencias fotográficas")
}

// ==========================================
// 8. Extreme Out of Bounds GPS coordinates (Invalid geographic locations)
// ==========================================
func TestQAUserErrors_ExtremeGPSCoordinates(t *testing.T) {
	s := &services.AttendanceService{}
	center := models.WorkCenter{
		Latitude:              19.4326,
		Longitude:             -99.1332,
		ToleranceRadiusMeters: 100,
	}

	// Using math.NaN() to represent completely invalid/malformed float inputs
	distNaN := s.CalculateDistance(math.NaN(), math.NaN(), center.Latitude, center.Longitude)

	// Confirm that the result is NaN (math.IsNaN) rather than panicking
	assert.True(t, math.IsNaN(distNaN))
}

