package services

import (
	"attendance-api/internal/models"
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// ==========================================
// 1. Midnight-Crossing Shifts (Night Shifts)
// ==========================================

func TestQANightShift_CheckIfLate(t *testing.T) {
	s := &AttendanceService{}
	shift := models.WorkShift{
		ExpectedCheckIn: "22:00:00",
		ExpectedCheckOut: "06:00:00",
		ToleranceTime:   "00:15:00",
		IsNightShift:     true,
	}
	timezone := "UTC"

	// Case 1.1: On-time check-in before shift start (21:55:00)
	checkInOnTime1 := time.Date(2026, 6, 6, 21, 55, 0, 0, time.UTC)
	isLate, delay := s.CheckIfLate(checkInOnTime1, shift, timezone)
	assert.False(t, isLate, "Expected to be on-time at 21:55")
	assert.Equal(t, 0, delay)

	// Case 1.2: On-time check-in within grace period (22:10:00)
	checkInOnTime2 := time.Date(2026, 6, 6, 22, 10, 0, 0, time.UTC)
	isLate, delay = s.CheckIfLate(checkInOnTime2, shift, timezone)
	assert.False(t, isLate, "Expected to be on-time within 15 mins grace period")
	assert.Equal(t, 0, delay)

	// Case 1.3: Late check-in beyond grace period (22:20:00)
	checkInLate := time.Date(2026, 6, 6, 22, 20, 0, 0, time.UTC)
	isLate, delay = s.CheckIfLate(checkInLate, shift, timezone)
	assert.True(t, isLate, "Expected to be late at 22:20")
	assert.Equal(t, 20, delay, "Delay should be calculated from shift start, not from grace limit")
}

func TestQANightShift_IsTooEarly(t *testing.T) {
	s := &AttendanceService{}
	shift := models.WorkShift{
		ExpectedCheckIn: "22:00:00",
		IsNightShift:     true,
	}
	timezone := "UTC"

	// Case 1.4: Too early (more than 2 hours before, e.g. 19:30:00)
	checkInTooEarly := time.Date(2026, 6, 6, 19, 30, 0, 0, time.UTC)
	tooEarly := s.IsTooEarly(checkInTooEarly, shift, timezone)
	assert.True(t, tooEarly, "Expected 19:30 to be too early for a 22:00 shift")

	// Case 1.5: Not too early (within 2 hours, e.g. 20:30:00)
	checkInOK := time.Date(2026, 6, 6, 20, 30, 0, 0, time.UTC)
	tooEarly = s.IsTooEarly(checkInOK, shift, timezone)
	assert.False(t, tooEarly, "Expected 20:30 NOT to be too early for a 22:00 shift")
}

func TestQANightShift_IsShiftOver(t *testing.T) {
	s := &AttendanceService{}
	shift := models.WorkShift{
		ExpectedCheckIn:  "22:00:00",
		ExpectedCheckOut: "06:00:00",
		IsNightShift:     true,
	}
	timezone := "UTC"

	// Case 1.6: Shift NOT over at 05:30:00 on day 2
	timeDuringShift := time.Date(2026, 6, 7, 5, 30, 0, 0, time.UTC)
	isOver := s.IsShiftOver(timeDuringShift, shift, timezone)
	assert.False(t, isOver, "Expected shift not to be over at 05:30")

	// Case 1.7: Shift IS over at 07:00:00 on day 2
	timeAfterShift := time.Date(2026, 6, 7, 7, 0, 0, 0, time.UTC)
	isOver = s.IsShiftOver(timeAfterShift, shift, timezone)
	assert.True(t, isOver, "Expected shift to be over at 07:00")
}

// ==========================================
// 2. Missing/Nil Data Handling
// ==========================================

func TestQAMissingData_RecalculateAttendance_NilTimestamps(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening mock db: %s", err)
	}
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")
	s := &AttendanceService{}

	// Case 2.1: Nil CheckOut time (incomplete session)
	checkIn := time.Now()
	att := models.Attendance{
		ID:         1,
		EmployeeID: 10,
		CheckIn:    &checkIn,
		CheckOut:   nil, // Faltante
	}

	mock.ExpectQuery("SELECT \\* FROM attendances").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "employee_id", "check_in", "check_out"}).
			AddRow(att.ID, att.EmployeeID, att.CheckIn, att.CheckOut))

	err = s.RecalculateAttendance(context.Background(), sqlxDB, 1)
	assert.NoError(t, err, "RecalculateAttendance should return nil (no-op) and not panic if CheckOut is nil")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQAMissingData_RecalculateAttendance_NilLunch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening mock db: %s", err)
	}
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")
	s := &AttendanceService{}

	// Case 2.2: Complete timestamps but missing/nil lunch times (no-lunch taken)
	checkIn := time.Date(2026, 6, 6, 8, 0, 0, 0, time.UTC)
	checkOut := time.Date(2026, 6, 6, 16, 0, 0, 0, time.UTC) // 8 hours duration
	att := models.Attendance{
		ID:         2,
		EmployeeID: 10,
		CheckIn:    &checkIn,
		CheckOut:   &checkOut,
		LunchStart: nil, // Faltante
		LunchEnd:   nil, // Faltante
	}

	mock.ExpectQuery("SELECT \\* FROM attendances WHERE id = \\$1").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "employee_id", "check_in", "check_out", "lunch_start", "lunch_end"}).
			AddRow(att.ID, att.EmployeeID, att.CheckIn, att.CheckOut, att.LunchStart, att.LunchEnd))

	mock.ExpectQuery("SELECT \\* FROM employees WHERE id = \\$1").
		WithArgs(att.EmployeeID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "position_id"}).AddRow(att.EmployeeID, 5))

	mock.ExpectQuery("SELECT \\* FROM positions WHERE id = \\$1").
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "hourly_rate"}).AddRow(5, 150.0))

	mock.ExpectQuery("SELECT \\* FROM incidents WHERE attendance_id = \\$1").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}))

	// Should calculate 8.0 hours worked and 8 * 150 = 1200.0 earnings
	mock.ExpectExec("UPDATE attendances SET").
		WithArgs(8.0, 1200.0, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = s.RecalculateAttendance(context.Background(), sqlxDB, 2)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==========================================
// 3. Grace Period Boundaries
// ==========================================

func TestQAGracePeriod_CheckIfLate_Boundaries(t *testing.T) {
	s := &AttendanceService{}
	shift := models.WorkShift{
		ExpectedCheckIn: "09:00:00",
		ToleranceTime:   "00:10:00",
		IsNightShift:     false,
	}
	timezone := "UTC"

	// Case 3.1: Exactly at the grace limit (09:10:00) -> NOT late
	checkInOnTimeLimit := time.Date(2026, 6, 6, 9, 10, 0, 0, time.UTC)
	isLate, delay := s.CheckIfLate(checkInOnTimeLimit, shift, timezone)
	assert.False(t, isLate, "Exactly at grace limit (09:10:00) should be on-time")
	assert.Equal(t, 0, delay)

	// Case 3.2: 1 second past grace limit (09:10:01) -> Late
	checkInLateLimit := time.Date(2026, 6, 6, 9, 10, 1, 0, time.UTC)
	isLate, delay = s.CheckIfLate(checkInLateLimit, shift, timezone)
	assert.True(t, isLate, "1 second past grace limit (09:10:01) should be late")
	assert.Equal(t, 10, delay) // delay in minutes rounded/truncated

	// Case 3.3: 1 minute past grace limit (09:11:00) -> Late
	checkInLateMinutes := time.Date(2026, 6, 6, 9, 11, 0, 0, time.UTC)
	isLate, delay = s.CheckIfLate(checkInLateMinutes, shift, timezone)
	assert.True(t, isLate, "09:11:00 is late")
	assert.Equal(t, 11, delay)
}

// ==========================================
// 4. Geofencing Proximity Boundaries
// ==========================================

func TestQAGeofence_BoundaryChecks(t *testing.T) {
	s := &AttendanceService{}
	center := models.WorkCenter{
		Latitude:              19.4326,
		Longitude:             -99.1332,
		ToleranceRadiusMeters: 100,
	}

	// Case 4.1: Within boundary (approx 94.5m)
	// 19.4326, -99.1332 vs 19.43345, -99.1332
	distWithin := s.CalculateDistance(center.Latitude, center.Longitude, 19.43345, center.Longitude)
	assert.True(t, distWithin <= float64(center.ToleranceRadiusMeters), "Distance should be under 100m")

	// Case 4.2: Outside boundary (approx 106.7m)
	// 19.4326, -99.1332 vs 19.43356, -99.1332
	distOutside := s.CalculateDistance(center.Latitude, center.Longitude, 19.43356, center.Longitude)
	assert.True(t, distOutside > float64(center.ToleranceRadiusMeters), "Distance should be over 100m")
}

// ==========================================
// 5. Shift Policy Types (Flexible / Field / Fixed)
// ==========================================

func TestQAShiftPolicies_AutoDetectIncidents(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening mock db: %s", err)
	}
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")
	s := &AttendanceService{}

	// Case 5.1: Flexible Shift (flexible) -> ignores tardiness (no late incident created)
	checkInTime := time.Date(2026, 6, 6, 11, 0, 0, 0, time.UTC) // 2 hours late
	attFlexible := models.Attendance{
		ID:          10,
		EmployeeID:  100,
		WorkShiftID: intPtr(1),
		CheckIn:     &checkInTime,
	}
	shiftFlexible := models.WorkShift{
		ID:              1,
		ExpectedCheckIn: "09:00:00",
		ToleranceTime:   "00:15:00",
		EnforceLateness: true,
		ShiftType:       "flexible", // flexible policy!
	}

	mock.ExpectQuery("SELECT \\* FROM attendances WHERE id = \\$1").
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "employee_id", "work_shift_id", "check_in"}).
			AddRow(attFlexible.ID, attFlexible.EmployeeID, attFlexible.WorkShiftID, attFlexible.CheckIn))

	mock.ExpectQuery("SELECT \\* FROM work_shifts WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "expected_check_in", "tolerance_time", "enforce_lateness", "shift_type"}).
			AddRow(shiftFlexible.ID, shiftFlexible.ExpectedCheckIn, shiftFlexible.ToleranceTime, shiftFlexible.EnforceLateness, shiftFlexible.ShiftType))

	mock.ExpectQuery("SELECT \\* FROM employees WHERE id = \\$1").
		WithArgs(attFlexible.EmployeeID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "position_id"}).AddRow(attFlexible.EmployeeID, 3))

	mock.ExpectExec("DELETE FROM incidents WHERE attendance_id = \\$1").
		WithArgs(10, string(models.IncidentTypeLunchOverstay), string(models.StatusPending)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Under RecalculateAttendance:
	mock.ExpectQuery("SELECT \\* FROM attendances WHERE id = \\$1").
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "employee_id", "work_shift_id", "check_in"}).
			AddRow(attFlexible.ID, attFlexible.EmployeeID, attFlexible.WorkShiftID, attFlexible.CheckIn))

	err = s.AutoDetectIncidents(context.Background(), sqlxDB, 10)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func intPtr(i int) *int {
	return &i
}
