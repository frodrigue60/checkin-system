package services

import (
	"attendance-api/internal/models"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type AuditService struct {
	DB *sqlx.DB
}

func (s *AuditService) LogAction(userID int, action string, entityType string, entityID int, oldValue interface{}, newValue interface{}, ipAddress string) error {
	return s.LogActionTx(nil, userID, action, entityType, entityID, oldValue, newValue, ipAddress)
}

func (s *AuditService) LogActionTx(tx *sqlx.Tx, userID int, action string, entityType string, entityID int, oldValue interface{}, newValue interface{}, ipAddress string) error {
	var oldJSON, newJSON *string

	if oldValue != nil {
		b, err := json.Marshal(oldValue)
		if err == nil {
			s := string(b)
			oldJSON = &s
		}
	}

	if newValue != nil {
		b, err := json.Marshal(newValue)
		if err == nil {
			s := string(b)
			newJSON = &s
		}
	}

	query := `
		INSERT INTO audit_logs (user_id, action, entity_type, entity_id, old_value, new_value, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	var err error
	if tx != nil {
		_, err = tx.Exec(query, userID, action, entityType, entityID, oldJSON, newJSON, ipAddress)
	} else {
		_, err = s.DB.Exec(query, userID, action, entityType, entityID, oldJSON, newJSON, ipAddress)
	}

	if err != nil {
		fmt.Printf("Error logging audit: %v\n", err)
	}

	return err
}

func (s *AuditService) ListAuditLogs(limit, offset int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	query := `
		SELECT l.*, u.name as user_name 
		FROM audit_logs l
		LEFT JOIN users u ON l.user_id = u.id
		ORDER BY l.created_at DESC
		LIMIT $1 OFFSET $2
	`
	err := s.DB.Select(&logs, query, limit, offset)
	return logs, err
}
