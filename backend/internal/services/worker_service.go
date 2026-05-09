package services

import (
	"attendance-api/internal/models"
	"attendance-api/internal/utils"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
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
	utils.GetLogger().Info("Ghost Session Cleaner started (every 30m)")
}

func (s *WorkerService) AutoCloseSessions() {
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	utils.GetLogger().Info("Running AutoCloseSessions...", zap.String("time", now.Format(time.RFC3339)))

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
	err := s.DB.SelectContext(ctx, &openAttendances, query)
	if err != nil {
		utils.GetLogger().Error("Error fetching ghost sessions", zap.Error(err))
		return
	}

	for _, att := range openAttendances {
		utils.GetLogger().Info("Auto-closing ghost session", zap.Int("employee_id", att.EmployeeID), zap.Int("attendance_id", att.ID))
		
		// Set check_out to NOW or to Expected End?
		// Proposed: Set to Expected End to avoid inflating hours, 
		// but maybe NOW is more realistic for "when we caught it".
		// Business rule says "close after 3 hours", let's set it to the time it SHOULD have ended.
		
		_, err := s.DB.ExecContext(ctx, `
			UPDATE attendances 
			SET check_out = (check_in::date + (SELECT expected_check_out FROM work_shifts WHERE id = (SELECT work_shift_id FROM employees WHERE id = $1))),
			    is_incomplete = true,
				updated_at = NOW()
			WHERE id = $2`, 
			att.EmployeeID, att.ID)
		
		if err != nil {
			utils.GetLogger().Error("Error auto-closing session", zap.Int("attendance_id", att.ID), zap.Error(err))
			continue
		}

		// Recalculate earnings
		s.AttendanceService.RecalculateAttendance(ctx, s.DB, att.ID)

		// Create Alert
		msg := fmt.Sprintf("Sesión olvidada cerrada automáticamente para el empleado ID %d (Asistencia #%d)", att.EmployeeID, att.ID)
		s.AlertService.CreateAlert(ctx, s.DB, "ghost_session_closed", "warning", msg, map[string]interface{}{
			"attendance_id": att.ID,
			"employee_id":   att.EmployeeID,
		})
	}
}
