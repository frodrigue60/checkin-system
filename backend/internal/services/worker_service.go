package services

import (
	"attendance-api/internal/models"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type WorkerService struct {
	DB                *sqlx.DB
	AttendanceService *AttendanceService
	AlertService      *AlertService
}

func (s *WorkerService) StartGhostSessionCleaner() {
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		for range ticker.C {
			s.AutoCloseSessions()
		}
	}()
	fmt.Println("Ghost Session Cleaner started (every 30m)")
}

func (s *WorkerService) AutoCloseSessions() {
	now := time.Now()
	fmt.Printf("[%s] Running AutoCloseSessions...\n", now.Format(time.RFC3339))

	// Find open sessions (check_out is null) where shift ended > 3 hours ago
	// We join with work_shifts to get the expected end time.
	// We use the date of check_in and the expected_check_out time.
	
	query := `
		SELECT a.* 
		FROM attendances a
		JOIN employees e ON a.employee_id = e.id
		JOIN work_shifts s ON e.work_shift_id = s.id
		WHERE a.check_out IS NULL 
		AND a.is_absence = false
		AND (a.check_in::date + s.expected_check_out) < (NOW() - INTERVAL '3 hours')
	`
	
	var openAttendances []models.Attendance
	err := s.DB.Select(&openAttendances, query)
	if err != nil {
		fmt.Printf("Error fetching ghost sessions: %v\n", err)
		return
	}

	for _, att := range openAttendances {
		fmt.Printf("Auto-closing ghost session for Employee %d (Attendance %d)\n", att.EmployeeID, att.ID)
		
		// Set check_out to NOW or to Expected End?
		// Proposed: Set to Expected End to avoid inflating hours, 
		// but maybe NOW is more realistic for "when we caught it".
		// Business rule says "close after 3 hours", let's set it to the time it SHOULD have ended.
		
		_, err := s.DB.Exec(`
			UPDATE attendances 
			SET check_out = (check_in::date + (SELECT expected_check_out FROM work_shifts WHERE id = (SELECT work_shift_id FROM employees WHERE id = $1))),
			    is_incomplete = true,
				updated_at = NOW()
			WHERE id = $2`, 
			att.EmployeeID, att.ID)
		
		if err != nil {
			fmt.Printf("Error auto-closing session %d: %v\n", att.ID, err)
			continue
		}

		// Recalculate earnings
		s.AttendanceService.RecalculateAttendance(s.DB, att.ID)

		// Create Alert
		msg := fmt.Sprintf("Sesión olvidada cerrada automáticamente para el empleado ID %d (Asistencia #%d)", att.EmployeeID, att.ID)
		s.AlertService.CreateAlert("ghost_session_closed", "warning", msg, map[string]interface{}{
			"attendance_id": att.ID,
			"employee_id":   att.EmployeeID,
		})
	}
}
