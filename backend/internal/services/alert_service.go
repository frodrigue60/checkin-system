package services

import (
	"attendance-api/internal/models"
	"encoding/json"
	"github.com/jmoiron/sqlx"
)

type AlertService struct {
	DB *sqlx.DB
}

func (s *AlertService) CreateAlert(alertType, severity, message string, metadata interface{}) error {
	return s.CreateAlertTx(nil, alertType, severity, message, metadata)
}

func (s *AlertService) CreateAlertTx(tx *sqlx.Tx, alertType, severity, message string, metadata interface{}) error {
	var metadataStr *string
	if metadata != nil {
		b, err := json.Marshal(metadata)
		if err == nil {
			s := string(b)
			metadataStr = &s
		}
	}

	query := `
		INSERT INTO system_alerts (type, severity, message, metadata_json)
		VALUES ($1, $2, $3, $4)`

	var err error
	if tx != nil {
		_, err = tx.Exec(query, alertType, severity, message, metadataStr)
	} else {
		_, err = s.DB.Exec(query, alertType, severity, message, metadataStr)
	}
	
	return err
}

func (s *AlertService) ListAlerts(onlyUnread bool, limit int) ([]models.SystemAlert, error) {
	alerts := []models.SystemAlert{}
	query := "SELECT * FROM system_alerts "
	if onlyUnread {
		query += "WHERE is_read = false "
	}
	query += "ORDER BY created_at DESC LIMIT $1"
	
	err := s.DB.Select(&alerts, query, limit)
	return alerts, err
}

func (s *AlertService) MarkAsRead(id int) error {
	_, err := s.DB.Exec("UPDATE system_alerts SET is_read = true WHERE id = $1", id)
	return err
}
