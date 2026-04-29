package services

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestReportService_InvalidateReports(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")

	s := &ReportService{DB: sqlxDB}

	t.Run("Successfully invalidate overlapping reports", func(t *testing.T) {
		employeeID := 1
		attendanceDate := time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC)
		dateStr := "2026-04-24"

		mock.ExpectExec("UPDATE reports SET status = 'stale'").
			WithArgs(employeeID, dateStr).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := s.InvalidateReports(employeeID, attendanceDate)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
