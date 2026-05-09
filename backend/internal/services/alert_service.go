package services

import (
	"attendance-api/internal/database"
	"attendance-api/internal/models"
	"context"
	"encoding/json"
	"github.com/jmoiron/sqlx"
)

type AlertService struct {
	DB *sqlx.DB
}

func (s *AlertService) CreateAlert(ctx context.Context, q database.Querier, alertType, severity, message string, metadata interface{}) error {
	var metadataStr *string
	if metadata != nil {
		b, err := json.Marshal(metadata)
		if err == nil {
			str := string(b)
			metadataStr = &str
		}
	}

	query := `
		INSERT INTO system_alerts (type, severity, message, metadata_json)
		VALUES ($1, $2, $3, $4)`

	var err error
	if q != nil {
		_, err = q.ExecContext(ctx, query, alertType, severity, message, metadataStr)
	} else {
		_, err = s.DB.ExecContext(ctx, query, alertType, severity, message, metadataStr)
	}
	
	return err
}

func (s *AlertService) ListAlerts(ctx context.Context, onlyUnread bool, limit int) ([]models.SystemAlert, error) {
	alerts := []models.SystemAlert{}
	query := "SELECT * FROM system_alerts "
	if onlyUnread {
		query += "WHERE is_read = false "
	}
	query += "ORDER BY created_at DESC LIMIT $1"
	
	err := s.DB.SelectContext(ctx, &alerts, query, limit)
	return alerts, err
}

func (s *AlertService) MarkAsRead(ctx context.Context, id int) error {
	_, err := s.DB.ExecContext(ctx, "UPDATE system_alerts SET is_read = true WHERE id = $1", id)
	return err
}
