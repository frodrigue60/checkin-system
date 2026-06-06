package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMapUserToDTO(t *testing.T) {
	now := time.Now()
	u := User{
		ID:        1,
		Name:      "Test User",
		Email:     "test@test.com",
		Password:  "SECRET_HASH",
		RoleID:    2,
		CreatedAt: &now,
	}

	dto := MapUserToDTO(u, "http://localhost:8080")

	assert.Equal(t, u.ID, dto.ID)
	assert.Equal(t, u.Name, dto.Name)
	assert.Equal(t, u.Email, dto.Email)
	assert.Equal(t, u.RoleID, dto.RoleID)
	assert.Equal(t, now.Format(time.RFC3339), dto.CreatedAt)
	
	// Ensure Password is NOT in DTO (compile time check mostly, but also value check)
	// UserDTO doesn't have a Password field, so we can't even access it.
}

func TestMapAttendanceToDTO(t *testing.T) {
	checkIn := time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC)
	
	// Case 1: Complete record
	hours := 8.5
	earnings := 850.0
	a := Attendance{
		ID:             1,
		EmployeeID:     10,
		CheckIn:        &checkIn,
		NetHoursWorked: &hours,
		DailyEarnings:  &earnings,
		IsAbsence:      false,
	}

	dto := MapAttendanceToDTO(a, "http://localhost:8080")
	assert.Equal(t, "2026-04-24", dto.Date)
	assert.Equal(t, 8.5, dto.NetHoursWorked)
	assert.Equal(t, 850.0, dto.DailyEarnings)

	// Case 2: Nil pointers
	aNil := Attendance{
		ID:             2,
		EmployeeID:     10,
		CheckIn:        &checkIn,
		NetHoursWorked: nil,
		DailyEarnings:  nil,
	}
	dtoNil := MapAttendanceToDTO(aNil, "http://localhost:8080")
	assert.Equal(t, 0.0, dtoNil.NetHoursWorked)
	assert.Equal(t, 0.0, dtoNil.DailyEarnings)
}

func TestMapReportToDTO(t *testing.T) {
	hoursStr := "8.0"
	earningsStr := "800.0"
	breakdown := []DailyBreakdownItem{
		{Date: "2026-04-24", NetHours: &hoursStr, Earnings: &earningsStr},
	}
	breakdownJSON, _ := json.Marshal(breakdown)

	r := Report{
		ID:               1,
		EmployeeID:       10,
		StartDate:        time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		EndDate:          time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		TotalHoursWorked: 160.0,
		DailyBreakdown:   breakdownJSON,
	}

	dto := MapReportToDTO(r)
	assert.Equal(t, 1, len(dto.DailyBreakdown))
	assert.Equal(t, "2026-04-24", dto.DailyBreakdown[0].Date)
	assert.Equal(t, "8.0", *dto.DailyBreakdown[0].NetHours)
}
