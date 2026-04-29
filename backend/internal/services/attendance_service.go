package services

import (
	"attendance-api/internal/models"
	"fmt"
	"math"
	"time"

	"github.com/jmoiron/sqlx"
)

type AttendanceService struct{}

// CalculateDistance returns the distance between two points in meters using the Haversine formula
func (s *AttendanceService) CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth radius in meters
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	deltaPhi := (lat2 - lat1) * math.Pi / 180
	deltaLambda := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*
			math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// CheckIfLate checks if the check-in time is beyond the shift start time + grace period
// Returns (isLate, delayMinutes)
func (s *AttendanceService) CheckIfLate(checkInTime time.Time, shift models.WorkShift, timezone string) (bool, int) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	// Localize the check-in time to the WorkCenter's timezone
	localTime := checkInTime.In(loc)

	startTime, err := s.ParseFlexibleTime(shift.ExpectedCheckIn)
	if err != nil {
		return false, 0
	}

	sh, sm, ss := startTime.Clock()
	// Target start time on the SAME day as check-in
	targetStart := time.Date(localTime.Year(), localTime.Month(), localTime.Day(), sh, sm, ss, 0, loc)

	// Adjust for night shift if necessary
	if shift.IsNightShift {
		// If checking in very early (e.g. 0-12 AM) for a shift that starts late (e.g. 10 PM),
		// it might be that targetStart should be YESTERDAY.
		if localTime.Hour() < 12 && sh >= 12 {
			targetStart = targetStart.Add(-24 * time.Hour)
		}
		// Or if checking in late (e.g. 10 PM) for a shift that starts early (e.g. 2 AM),
		// it might be that targetStart should be TOMORROW.
		if localTime.Hour() >= 12 && sh < 12 {
			targetStart = targetStart.Add(24 * time.Hour)
		}
	}

	toleranceTime, err := s.ParseFlexibleTime(shift.ToleranceTime)
	tolMinutes := 0
	if err == nil {
		th, tm, _ := toleranceTime.Clock()
		tolMinutes = th*60 + tm
	}

	limitTime := targetStart.Add(time.Duration(tolMinutes) * time.Minute)
	isLate := localTime.After(limitTime)
	delayMinutes := 0
	if isLate {
		delayMinutes = int(localTime.Sub(targetStart).Minutes())
	}

	return isLate, delayMinutes
}

// IsShiftOver checks if the current time is beyond the expected checkout time
func (s *AttendanceService) IsShiftOver(checkInTime time.Time, shift models.WorkShift, timezone string) bool {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	localTime := checkInTime.In(loc)
	endTime, err := s.ParseFlexibleTime(shift.ExpectedCheckOut)
	if err != nil {
		return false
	}

	eh, em, es := endTime.Clock()
	targetEnd := time.Date(localTime.Year(), localTime.Month(), localTime.Day(), eh, em, es, 0, loc)

	if shift.IsNightShift {
		startTime, _ := s.ParseFlexibleTime(shift.ExpectedCheckIn)
		sh, _, _ := startTime.Clock()
		
		// If end clock is smaller than start clock, it's next day
		if eh < sh {
			// If we are currently "before" the start time of the day, 
			// then targetEnd is indeed today.
			// But if we are "after" the start time, the shift hasn't ended yet (it ends tomorrow).
			if localTime.Hour() >= sh {
				targetEnd = targetEnd.Add(24 * time.Hour)
			}
		}
	}

	return localTime.After(targetEnd)
}

// GetShiftOverDuration returns the hours and minutes passed since the shift end
func (s *AttendanceService) GetShiftOverDuration(checkInTime time.Time, shift models.WorkShift, timezone string) (int, int) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	localTime := checkInTime.In(loc)
	endTime, err := s.ParseFlexibleTime(shift.ExpectedCheckOut)
	if err != nil {
		return 0, 0
	}

	h, m, s_ := localTime.Clock()
	checkInSeconds := h*3600 + m*60 + s_

	eh, em, es := endTime.Clock()
	endSeconds := eh*3600 + em*60 + es

	diffSeconds := checkInSeconds - endSeconds
	if diffSeconds < 0 {
		return 0, 0
	}

	hours := diffSeconds / 3600
	minutes := (diffSeconds % 3600) / 60
	return hours, minutes
}

// IsTooEarly checks if the check-in is more than 2 hours before shift start
func (s *AttendanceService) IsTooEarly(checkInTime time.Time, shift models.WorkShift, timezone string) bool {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	localTime := checkInTime.In(loc)
	startTime, err := s.ParseFlexibleTime(shift.ExpectedCheckIn)
	if err != nil {
		return false
	}

	sh, sm, ss := startTime.Clock()
	targetStart := time.Date(localTime.Year(), localTime.Month(), localTime.Day(), sh, sm, ss, 0, loc)

	if shift.IsNightShift {
		// Similar adjustment as CheckIfLate
		if localTime.Hour() < 12 && sh >= 12 {
			targetStart = targetStart.Add(-24 * time.Hour)
		}
		if localTime.Hour() >= 12 && sh < 12 {
			targetStart = targetStart.Add(24 * time.Hour)
		}
	}

	return localTime.Before(targetStart.Add(-2 * time.Hour))
}

// ParseFlexibleTime handles multiple time formats gracefully
func (s *AttendanceService) ParseFlexibleTime(tStr string) (time.Time, error) {
	if tStr == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}

	formats := []string{
		"15:04:05",
		"15:04",
		time.RFC3339,
		"2006-01-02T15:04:05Z",
	}

	for _, f := range formats {
		t, err := time.Parse(f, tStr)
		if err == nil {
			return t, nil
		}
	}

	// Fallback for Laravel-style fractional seconds or weird ISO variants
	if len(tStr) > 19 {
		t, err := time.Parse("2006-01-02T15:04:05", tStr[:19])
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse time: %s", tStr)
}


// CalculateEarnings computes the daily earnings based on hours worked and penalties
func (s *AttendanceService) CalculateEarnings(hours float64, hourlyRate float64, incidents []models.Incident, pos models.Position) float64 {
	gross := hours * hourlyRate
	deductions := 0.0

	for _, incident := range incidents {
		// Skip justified incidents
		if incident.Status == models.StatusJustified {
			continue
		}

		if incident.Type == models.IncidentTypeLate {
			deductions += pos.LatePenaltyFee
		} else if incident.Type == models.IncidentTypeOutOfRange {
			deductions += pos.OutOfRangeFee
		} else if incident.Type == models.IncidentTypeLunchOverstay {
			deductions += pos.LunchOverstayFee
		}
	}

	net := gross - deductions
	if net < 0 {
		return 0
	}
	// Round to 2 decimal places for financial precision
	return math.Round(net*100) / 100
}

// RecalculateAttendance updates the net hours and earnings based on current timestamps and incidents
func (s *AttendanceService) RecalculateAttendance(db *sqlx.DB, attID int) error {
	var att models.Attendance
	if err := db.Get(&att, "SELECT * FROM attendances WHERE id = $1", attID); err != nil {
		return err
	}

	if att.CheckIn == nil || att.CheckOut == nil {
		return nil // Cannot recalculate incomplete session
	}

	// 1. Calculate Duration
	duration := att.CheckOut.Sub(*att.CheckIn)
	
	// Subtract lunch if exists
	if att.LunchStart != nil && att.LunchEnd != nil {
		lunchDuration := att.LunchEnd.Sub(*att.LunchStart)
		duration -= lunchDuration
	}
	
	hours := duration.Hours()
	if hours < 0 { hours = 0 }

	// 2. Get Employee & Position
	var emp models.Employee
	if err := db.Get(&emp, "SELECT * FROM employees WHERE id = $1", att.EmployeeID); err != nil {
		return err
	}

	var pos models.Position
	if err := db.Get(&pos, "SELECT * FROM positions WHERE id = $1", emp.PositionID); err != nil {
		return err
	}

	// 3. Get Incidents
	var incidents []models.Incident
	if err := db.Unsafe().Select(&incidents, "SELECT * FROM incidents WHERE attendance_id = $1", att.ID); err != nil {
		return err
	}

	// 4. Calculate Earnings
	earnings := s.CalculateEarnings(hours, pos.HourlyRate, incidents, pos)

	// 5. Update DB
	_, err := db.Exec("UPDATE attendances SET net_hours_worked = $1, daily_earnings = $2, updated_at = NOW() WHERE id = $3", hours, earnings, att.ID)
	return err
}

// CreateGeofenceIncident creates an out of range incident if necessary
func (s *AttendanceService) CreateGeofenceIncident(tx *sqlx.Tx, empID, centerID, attID int, lat, lon float64, center models.WorkCenter, action string) error {
	distance := s.CalculateDistance(lat, lon, center.Latitude, center.Longitude)
	
	// If center is "Field/Mobile" (could be a specific ID or we check shift later)
	// For now, if center ID is 0 or dummy, we might skip. 
	// Better: The caller should decide based on shift rules.
	
	if distance > float64(center.ToleranceRadiusMeters) {
		now := time.Now()
		metadata := fmt.Sprintf(`{"action": "%s", "distance": %.2f, "limit": %d}`, action, distance, center.ToleranceRadiusMeters)
		_, err := tx.Exec(`INSERT INTO incidents 
			(employee_id, work_center_id, attendance_id, type, metadata_json, is_late, delay_minutes, is_out_of_range, distance, status, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`, 
			empID, centerID, attID, models.IncidentTypeOutOfRange, metadata, false, 0, true, int(distance), models.StatusPending, now, now)
		return err
	}
	return nil
}

// AutoDetectIncidentsTx is the transactional version of AutoDetectIncidents
func (s *AttendanceService) AutoDetectIncidentsTx(tx *sqlx.Tx, attID int) error {
	var att models.Attendance
	if err := tx.Get(&att, "SELECT * FROM attendances WHERE id = $1", attID); err != nil {
		return err
	}

	if att.WorkShiftID == nil {
		return nil
	}

	var shift models.WorkShift
	if err := tx.Get(&shift, "SELECT * FROM work_shifts WHERE id = $1", *att.WorkShiftID); err != nil {
		return err
	}

	var emp models.Employee
	if err := tx.Get(&emp, "SELECT * FROM employees WHERE id = $1", att.EmployeeID); err != nil {
		return err
	}

	var wc models.WorkCenter
	if att.WorkCenterID != nil {
		tx.Get(&wc, "SELECT * FROM work_centers WHERE id = $1", *att.WorkCenterID)
	}
	timezone := "UTC"
	if wc.Timezone != "" {
		timezone = wc.Timezone
	}

	now := time.Now()

	// 0. Check Absence
	if att.CheckIn == nil {
		if s.IsShiftOver(now, shift, timezone) {
			reason := "Inasistencia detectada automáticamente"
			_, err := tx.Exec(`UPDATE attendances SET is_absence = true, absence_reason = $1, updated_at = $2 WHERE id = $3`,
				reason, now, att.ID)
			if err != nil { return fmt.Errorf("error marking absence: %w", err) }
			att.IsAbsence = true
		}
	}

	// 1. Check Lateness
	if att.CheckIn != nil && !att.IsAbsence && shift.EnforceLateness && (shift.ShiftType == "fixed" || shift.ShiftType == "") {
		isLate, delayMinutes := s.CheckIfLate(*att.CheckIn, shift, timezone)
		if isLate {
			var count int
			tx.Get(&count, "SELECT COUNT(*) FROM incidents WHERE attendance_id = $1 AND type = $2", att.ID, models.IncidentTypeLate)
			if count == 0 {
				_, err := tx.Exec(`INSERT INTO incidents 
					(employee_id, work_center_id, attendance_id, type, delay_minutes, is_late, status, created_at, updated_at) 
					VALUES ($1, $2, $3, $4, $5, true, $6, $7, $8)`,
					att.EmployeeID, att.WorkCenterID, att.ID, models.IncidentTypeLate, delayMinutes, models.StatusPending, now, now)
				if err != nil { return fmt.Errorf("error inserting late incident: %w", err) }
			} else {
				tx.Exec("UPDATE incidents SET delay_minutes = $1, updated_at = NOW() WHERE attendance_id = $2 AND type = $3", delayMinutes, att.ID, models.IncidentTypeLate)
			}
		} else {
			tx.Exec("DELETE FROM incidents WHERE attendance_id = $1 AND type = $2 AND status = $3", att.ID, models.IncidentTypeLate, models.StatusPending)
		}
	} else if !shift.EnforceLateness {
		tx.Exec("DELETE FROM incidents WHERE attendance_id = $1 AND type = $2 AND status = $3", att.ID, models.IncidentTypeLate, models.StatusPending)
	}

	// 2. Check Lunch Overstay
	if att.LunchStart != nil && att.LunchEnd != nil && shift.EnforceLunchLimit {
		lunchDuration := att.LunchEnd.Sub(*att.LunchStart).Minutes()
		allowedTime, err := s.ParseFlexibleTime(shift.AllowedLunchTime)
		if err == nil {
			allowedMinutes := float64(allowedTime.Hour()*60 + allowedTime.Minute())
			if lunchDuration > (allowedMinutes + 1) { // 1 min grace
				delay := int(lunchDuration - allowedMinutes)
				var count int
				tx.Get(&count, "SELECT COUNT(*) FROM incidents WHERE attendance_id = $1 AND type = $2", att.ID, models.IncidentTypeLunchOverstay)
				if count == 0 {
					_, err := tx.Exec(`INSERT INTO incidents 
						(employee_id, work_center_id, attendance_id, type, delay_minutes, status, created_at, updated_at) 
						VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
						att.EmployeeID, att.WorkCenterID, att.ID, models.IncidentTypeLunchOverstay, delay, models.StatusPending, now, now)
					if err != nil { return fmt.Errorf("error inserting lunch incident: %w", err) }
				} else {
					tx.Exec("UPDATE incidents SET delay_minutes = $1, updated_at = NOW() WHERE attendance_id = $2 AND type = $3", delay, att.ID, models.IncidentTypeLunchOverstay)
				}
			} else {
				tx.Exec("DELETE FROM incidents WHERE attendance_id = $1 AND type = $2 AND status = $3", att.ID, models.IncidentTypeLunchOverstay, models.StatusPending)
			}
		}
	} else if !shift.EnforceLunchLimit {
		tx.Exec("DELETE FROM incidents WHERE attendance_id = $1 AND type = $2 AND status = $3", att.ID, models.IncidentTypeLunchOverstay, models.StatusPending)
	}

	return s.RecalculateAttendanceTx(tx, att.ID)
}

// RecalculateAttendanceTx is the transactional version of RecalculateAttendance
func (s *AttendanceService) RecalculateAttendanceTx(tx *sqlx.Tx, attID int) error {
	var att models.Attendance
	if err := tx.Get(&att, "SELECT * FROM attendances WHERE id = $1", attID); err != nil {
		return err
	}

	if att.CheckIn == nil || att.CheckOut == nil {
		return nil
	}

	duration := att.CheckOut.Sub(*att.CheckIn)
	if att.LunchStart != nil && att.LunchEnd != nil {
		duration -= att.LunchEnd.Sub(*att.LunchStart)
	}
	hours := duration.Hours()
	if hours < 0 { hours = 0 }

	var emp models.Employee
	if err := tx.Get(&emp, "SELECT * FROM employees WHERE id = $1", att.EmployeeID); err != nil {
		return err
	}

	var pos models.Position
	if err := tx.Get(&pos, "SELECT * FROM positions WHERE id = $1", emp.PositionID); err != nil {
		return err
	}

	var incidents []models.Incident
	if err := tx.Select(&incidents, "SELECT * FROM incidents WHERE attendance_id = $1", att.ID); err != nil {
		return err
	}

	earnings := s.CalculateEarnings(hours, pos.HourlyRate, incidents, pos)

	_, err := tx.Exec("UPDATE attendances SET net_hours_worked = $1, daily_earnings = $2, updated_at = NOW() WHERE id = $3", hours, earnings, att.ID)
	return err
}

// AutoDetectIncidents scans an attendance record and creates missing incidents based on shift rules
func (s *AttendanceService) AutoDetectIncidents(db *sqlx.DB, attID int) error {
	var att models.Attendance
	if err := db.Get(&att, "SELECT * FROM attendances WHERE id = $1", attID); err != nil {
		return err
	}

	if att.WorkShiftID == nil {
		return nil // No shift assigned, cannot detect incidents
	}

	var shift models.WorkShift
	if err := db.Get(&shift, "SELECT * FROM work_shifts WHERE id = $1", *att.WorkShiftID); err != nil {
		return err
	}

	var emp models.Employee
	if err := db.Get(&emp, "SELECT * FROM employees WHERE id = $1", att.EmployeeID); err != nil {
		return err
	}

	var wc models.WorkCenter
	if att.WorkCenterID != nil {
		db.Get(&wc, "SELECT * FROM work_centers WHERE id = $1", *att.WorkCenterID)
	}
	timezone := "UTC"
	if wc.Timezone != "" {
		timezone = wc.Timezone
	}

	now := time.Now()

	// 0. Check Absence (If no check-in and shift is over)
	if att.CheckIn == nil {
		if s.IsShiftOver(now, shift, timezone) {
			reason := "Inasistencia detectada automáticamente"
			_, err := db.Exec(`UPDATE attendances SET is_absence = true, absence_reason = $1, updated_at = $2 WHERE id = $3`,
				reason, now, att.ID)
			if err != nil { return fmt.Errorf("error marking absence: %w", err) }
			// Also update local copy for subsequent checks
			att.IsAbsence = true
			att.AbsenceReason = &reason
		}
	}

	// 1. Check Lateness
	if att.CheckIn != nil && !att.IsAbsence && shift.EnforceLateness && (shift.ShiftType == "fixed" || shift.ShiftType == "") {
		isLate, delayMinutes := s.CheckIfLate(*att.CheckIn, shift, timezone)
		if isLate {
			// Check if incident already exists
			var count int
			db.Get(&count, "SELECT COUNT(*) FROM incidents WHERE attendance_id = $1 AND type = $2", att.ID, models.IncidentTypeLate)
			if count == 0 {
				_, err := db.Exec(`INSERT INTO incidents 
					(employee_id, work_center_id, attendance_id, type, delay_minutes, is_late, status, created_at, updated_at) 
					VALUES ($1, $2, $3, $4, $5, true, $6, $7, $8)`,
					att.EmployeeID, att.WorkCenterID, att.ID, models.IncidentTypeLate, delayMinutes, models.StatusPending, now, now)
				if err != nil { return fmt.Errorf("error inserting late incident: %w", err) }
			} else {
				// Update delay if it changed
				db.Exec("UPDATE incidents SET delay_minutes = $1, updated_at = NOW() WHERE attendance_id = $2 AND type = $3", delayMinutes, att.ID, models.IncidentTypeLate)
			}
		} else {
			// No longer late? Delete existing late incident if it's still pending
			db.Exec("DELETE FROM incidents WHERE attendance_id = $1 AND type = $2 AND status = $3", att.ID, models.IncidentTypeLate, models.StatusPending)
		}
	} else if !shift.EnforceLateness {
		// If policy changed to non-enforced, remove pending incidents
		db.Exec("DELETE FROM incidents WHERE attendance_id = $1 AND type = $2 AND status = $3", att.ID, models.IncidentTypeLate, models.StatusPending)
	}

	// 2. Check Lunch Overstay
	if att.LunchStart != nil && att.LunchEnd != nil && shift.EnforceLunchLimit {
		lunchDuration := att.LunchEnd.Sub(*att.LunchStart).Minutes()
		
		allowedTime, err := time.Parse("15:04:05", shift.AllowedLunchTime)
		if err == nil {
			allowedMinutes := float64(allowedTime.Hour()*60 + allowedTime.Minute())
			if lunchDuration > allowedMinutes {
				delay := int(lunchDuration - allowedMinutes)
				var count int
				db.Get(&count, "SELECT COUNT(*) FROM incidents WHERE attendance_id = $1 AND type = $2", att.ID, models.IncidentTypeLunchOverstay)
				if count == 0 {
					_, err := db.Exec(`INSERT INTO incidents 
						(employee_id, work_center_id, attendance_id, type, delay_minutes, status, created_at, updated_at) 
						VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
						att.EmployeeID, att.WorkCenterID, att.ID, models.IncidentTypeLunchOverstay, delay, models.StatusPending, now, now)
					if err != nil { return fmt.Errorf("error inserting lunch incident: %w", err) }
				} else {
					// Update delay
					db.Exec("UPDATE incidents SET delay_minutes = $1, updated_at = NOW() WHERE attendance_id = $2 AND type = $3", delay, att.ID, models.IncidentTypeLunchOverstay)
				}
			} else {
				// No longer overstay? Delete if pending
				db.Exec("DELETE FROM incidents WHERE attendance_id = $1 AND type = $2 AND status = $3", att.ID, models.IncidentTypeLunchOverstay, models.StatusPending)
			}
		}
	} else if !shift.EnforceLunchLimit {
		// If policy changed, remove pending
		db.Exec("DELETE FROM incidents WHERE attendance_id = $1 AND type = $2 AND status = $3", att.ID, models.IncidentTypeLunchOverstay, models.StatusPending)
	}

	// 3. Recalculate financial data
	return s.RecalculateAttendance(db, att.ID)
}


