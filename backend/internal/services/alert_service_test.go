package services

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestAlertService_CreateAlert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error opening mock db: %s", err)
	}
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "postgres")

	s := &AlertService{DB: sqlxDB}

	t.Run("Successfully create alert with metadata", func(t *testing.T) {
		alertType := "test_alert"
		severity := "warn"
		message := "Something happened"
		metadata := map[string]int{"code": 123}

		mock.ExpectExec("INSERT INTO system_alerts").
			WithArgs(alertType, severity, message, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.CreateAlert(context.Background(), sqlxDB, alertType, severity, message, metadata)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
