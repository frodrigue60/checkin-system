package services

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestJustificationService_CreateJustification(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")

	alertService := &AlertService{DB: sqlxDB}
	s := &JustificationService{
		DB:           sqlxDB,
		AlertService: alertService,
	}

	t.Run("Successfully create justification", func(t *testing.T) {
		incidentID := 1
		employeeID := 10
		message := "Test justification"

		mock.ExpectBegin()

		// 1. Verify incident query
		rows := sqlmock.NewRows([]string{"id", "employee_id"}).
			AddRow(incidentID, employeeID)
		mock.ExpectQuery("SELECT \\* FROM incidents WHERE id = \\$1 AND employee_id = \\$2").
			WithArgs(incidentID, employeeID).
			WillReturnRows(rows)

		// 2. Check existing justification query
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM justifications WHERE incident_id = \\$1").
			WithArgs(incidentID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		// 3. Insert Justification
		mock.ExpectExec("INSERT INTO justifications").
			WithArgs(incidentID, employeeID, message, nil, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 4. Create Alert
		mock.ExpectExec("INSERT INTO system_alerts").
			WithArgs("justification_submitted", "info", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err := s.CreateJustification(incidentID, employeeID, message, nil)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Fail if incident not found", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery("SELECT \\* FROM incidents").
			WillReturnError(fmt.Errorf("not found"))

		mock.ExpectRollback()

		err := s.CreateJustification(1, 10, "msg", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "incidente no encontrado")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
