package services

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type ReportService struct {
	DB *sqlx.DB
}

// InvalidateReports marking existing reports as 'stale' if they overlap with a modified attendance date
func (s *ReportService) InvalidateReports(employeeID int, attendanceDate time.Time) error {
	// A report is affected if attendanceDate falls between start_date and end_date
	// We use check_in::date for the comparison
	dateStr := attendanceDate.Format("2006-01-02")
	
	query := `
		UPDATE reports 
		SET status = 'stale', updated_at = NOW() 
		WHERE employee_id = $1 
		AND $2 >= start_date AND $2 <= end_date
		AND status = 'valid'
	`
	
	_, err := s.DB.Exec(query, employeeID, dateStr)
	if err != nil {
		return fmt.Errorf("error invalidating reports: %w", err)
	}
	
	return nil
}
