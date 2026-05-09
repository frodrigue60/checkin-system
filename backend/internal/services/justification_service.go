package services

import (
	"attendance-api/internal/database"
	"attendance-api/internal/models"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type JustificationService struct {
	DB               *sqlx.DB
	AlertService     *AlertService
	AttendanceService *AttendanceService
}

func (s *JustificationService) CreateJustification(ctx context.Context, q database.Querier, incidentID int, employeeID int, message string, evidenceURL *string) error {
	// If q is provided, we use it directly (assuming the caller manages the transaction)
	// If q is nil, we manage our own transaction for backward compatibility or simple calls
	useTx := false
	var tx *sqlx.Tx
	var err error
	
	qi := q
	if qi == nil {
		tx, err = s.DB.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		qi = tx
		useTx = true
	}

	// 1. Verify incident belongs to employee
	var incident models.Incident
	err = qi.GetContext(ctx, &incident, "SELECT * FROM incidents WHERE id = $1 AND employee_id = $2", incidentID, employeeID)
	if err != nil {
		return errors.New("incidente no encontrado o no autorizado")
	}

	// 2. Check if already has a justification
	var count int
	qi.GetContext(ctx, &count, "SELECT COUNT(*) FROM justifications WHERE incident_id = $1", incidentID)
	if count > 0 {
		return errors.New("este incidente ya tiene una justificación en proceso")
	}

	// 3. Create Justification
	now := time.Now()
	_, err = qi.ExecContext(ctx, `
		INSERT INTO justifications (incident_id, employee_id, message, evidence_url, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'pending', $5, $5)`,
		incidentID, employeeID, message, evidenceURL, now)
	
	if err != nil {
		return fmt.Errorf("error al crear justificación: %v", err)
	}

	// 4. Create System Alert for Admin
	alertMsg := fmt.Sprintf("Nueva justificación pendiente del empleado ID %d para el incidente #%d", employeeID, incidentID)
	if err := s.AlertService.CreateAlert(ctx, qi, "justification_submitted", "info", alertMsg, map[string]interface{}{
		"incident_id":    incidentID,
		"employee_id":    employeeID,
		"justification": message,
	}); err != nil {
		return fmt.Errorf("error creating alert: %v", err)
	}

	if useTx {
		return tx.Commit()
	}
	return nil
}

func (s *JustificationService) ResolveJustification(ctx context.Context, q database.Querier, id int, adminID int, approve bool, note string) error {
	var just models.Justification
	
	qi := q
	if qi == nil { qi = s.DB }

	if err := qi.GetContext(ctx, &just, "SELECT * FROM justifications WHERE id = $1", id); err != nil {
		return errors.New("justificación no encontrada")
	}

	status := "rejected"
	incidentStatus := "pending"
	if approve {
		status = "approved"
		incidentStatus = "justified"
	}

	useInternalTx := false
	var tx *sqlx.Tx
	if q == nil {
		var err error
		tx, err = s.DB.BeginTxx(ctx, nil)
		if err != nil { return err }
		defer tx.Rollback()
		qi = tx
		useInternalTx = true
	}

	// Update Justification
	now := time.Now()
	_, err := qi.ExecContext(ctx, `
		UPDATE justifications 
		SET status = $1, resolved_by = $2, resolution_note = $3, updated_at = $4 
		WHERE id = $5`,
		status, adminID, note, now, id)
	if err != nil {
		return err
	}

	// Update Incident
	_, err = qi.ExecContext(ctx, `
		UPDATE incidents 
		SET status = $1, resolved_by = $2, resolution_note = $3, updated_at = $4 
		WHERE id = $5`,
		incidentStatus, adminID, note, now, just.IncidentID)
	if err != nil {
		return err
	}

	if useInternalTx {
		if err := tx.Commit(); err != nil {
			return err
		}
		// Reset qi to DB for the out-of-transaction recalculation if needed
		qi = s.DB
	}

	// Trigger Recalculation if approved
	if approve {
		var incident models.Incident
		qi.GetContext(ctx, &incident, "SELECT * FROM incidents WHERE id = $1", just.IncidentID)
		s.AttendanceService.RecalculateAttendance(ctx, qi, incident.AttendanceID)
	}

	return nil
}

func (s *JustificationService) ListPending(ctx context.Context, q database.Querier, limit int) ([]models.Justification, error) {
	var list []models.Justification
	qi := q
	if qi == nil { qi = s.DB }
	err := qi.SelectContext(ctx, &list, "SELECT * FROM justifications WHERE status = 'pending' ORDER BY created_at ASC LIMIT $1", limit)
	return list, err
}
