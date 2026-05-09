package services

import (
	"attendance-api/internal/database"
	"attendance-api/internal/models"
	"attendance-api/internal/utils"
	"context"
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type AuditService struct {
	DB *sqlx.DB
}

func (s *AuditService) LogAction(ctx context.Context, q database.Querier, userID int, action string, entityType string, entityID int, oldValue interface{}, newValue interface{}, ipAddress string) error {
	var oldJSON, newJSON *string

	if oldValue != nil {
		b, err := json.Marshal(oldValue)
		if err == nil {
			str := string(b)
			oldJSON = &str
		}
	}

	if newValue != nil {
		b, err := json.Marshal(newValue)
		if err == nil {
			str := string(b)
			newJSON = &str
		}
	}

	query := `
		INSERT INTO audit_logs (user_id, action, entity_type, entity_id, old_value, new_value, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	var err error
	if q != nil {
		_, err = q.ExecContext(ctx, query, userID, action, entityType, entityID, oldJSON, newJSON, ipAddress)
	} else {
		_, err = s.DB.ExecContext(ctx, query, userID, action, entityType, entityID, oldJSON, newJSON, ipAddress)
	}

	if err != nil {
		utils.GetLogger().Error("Error logging audit", zap.Error(err))
	}

	return err
}

func (s *AuditService) ListAuditLogs(ctx context.Context, limit, offset int) ([]models.AuditLog, error) {
	logs := []models.AuditLog{}
	query := `
		SELECT l.*, u.name as user_name 
		FROM audit_logs l
		LEFT JOIN users u ON l.user_id = u.id
		ORDER BY l.created_at DESC
		LIMIT $1 OFFSET $2
	`
	err := s.DB.SelectContext(ctx, &logs, query, limit, offset)
	return logs, err
}
