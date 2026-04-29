package services

import (
	"attendance-api/internal/models"
	"testing"
	"time"
)

func TestCalculateDistance(t *testing.T) {
	s := &AttendanceService{}
	// Test case: Same point
	dist := s.CalculateDistance(19.4326, -99.1332, 19.4326, -99.1332)
	if dist != 0 {
		t.Errorf("Expected 0 distance, got %f", dist)
	}

	// Test case: ~100m distance
	dist = s.CalculateDistance(19.4326, -99.1332, 19.4335, -99.1332)
	if dist < 90 || dist > 110 {
		t.Errorf("Expected approx 100m distance, got %f", dist)
	}
}

func TestCalculateEarnings(t *testing.T) {
	s := &AttendanceService{}
	pos := models.Position{
		HourlyRate:        100.0,
		LatePenaltyFee:    10.0,
		OutOfRangeFee:     20.0,
		LunchOverstayFee:  5.0,
	}

	incidents := []models.Incident{
		{Type: models.IncidentTypeLate, Status: models.StatusPending},
		{Type: models.IncidentTypeOutOfRange, Status: models.StatusPending},
		{Type: models.IncidentTypeLunchOverstay, Status: models.StatusJustified}, // Should be ignored
	}

	// 8 hours * 100/h = 800. Deductions: 10 + 20 = 30. Net: 770.
	earnings := s.CalculateEarnings(8.0, 100.0, incidents, pos)
	if earnings != 770.0 {
		t.Errorf("Expected 770.0 earnings, got %f", earnings)
	}
}

func TestCheckIfLate(t *testing.T) {
	s := &AttendanceService{}
	shift := models.WorkShift{
		ExpectedCheckIn: "08:00:00",
		ToleranceTime:   "00:15:00",
		EnforceLateness: true,
	}

	// Case 1: On time
	checkIn := time.Date(2026, 4, 25, 8, 10, 0, 0, time.UTC)
	isLate, _ := s.CheckIfLate(checkIn, shift, "UTC")
	if isLate {
		t.Errorf("Expected not late (8:10 < 8:15)")
	}

	// Case 2: Late
	checkIn = time.Date(2026, 4, 25, 8, 20, 0, 0, time.UTC)
	isLate, delay := s.CheckIfLate(checkIn, shift, "UTC")
	if !isLate {
		t.Errorf("Expected late (8:20 > 8:15)")
	}
	if delay != 20 {
		t.Errorf("Expected 20 min delay, got %d", delay)
	}
}
