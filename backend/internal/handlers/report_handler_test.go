package handlers

import (
	"attendance-api/internal/services"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestReportHandler_ListReports(t *testing.T) {
	app := fiber.New()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening mock db: %s", err)
	}
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")

	handler := &ReportHandler{DB: sqlxDB}
	app.Get("/reports", handler.ListReports)

	t.Run("Successfully list report ranges", func(t *testing.T) {
		startDate := time.Now().AddDate(0, 0, -30)
		endDate := time.Now()
		rows := sqlmock.NewRows([]string{"start_date", "end_date", "emp_count", "is_stale"}).
			AddRow(startDate, endDate, 10, false).
			AddRow(startDate.AddDate(0, 0, -30), startDate.AddDate(0, 0, -1), 12, true)

		mock.ExpectQuery("SELECT").WillReturnRows(rows)

		req := httptest.NewRequest("GET", "/reports", nil)
		resp, _ := app.Test(req)

		if resp.StatusCode != fiber.StatusOK {
			var errMap map[string]string
			// Fix: json is imported but I need to make sure I don't need a specific decoder if I can just use json.Unmarshal
			// Actually json.NewDecoder is fine.
			json.NewDecoder(resp.Body).Decode(&errMap)
			t.Errorf("expected 200, got %d. error: %s", resp.StatusCode, errMap["error"])
		}
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestReportHandler_DeleteReportByID(t *testing.T) {
	app := fiber.New()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening mock db: %s", err)
	}
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")

	auditService := &services.AuditService{DB: sqlxDB}
	handler := &ReportHandler{
		DB:           sqlxDB,
		AuditService: auditService,
	}
	
	// Mock middleware to set user_id
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", 1)
		return c.Next()
	})
	app.Delete("/reports/:id", handler.DeleteReportByID)

	t.Run("Successfully delete report", func(t *testing.T) {
		reportID := "123"

		mock.ExpectBegin()

		mock.ExpectExec("DELETE FROM reports WHERE id = \\$1").
			WithArgs(reportID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec("INSERT INTO audit_logs").
			WithArgs(1, "DELETE_REPORT", "report", 123, nil, nil, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		req := httptest.NewRequest("DELETE", "/reports/123", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}




