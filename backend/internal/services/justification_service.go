package services

import (
	"attendance-api/internal/models"
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

func (s *JustificationService) CreateJustification(incidentID int, employeeID int, message string, evidenceURL *string) error {
	tx, err := s.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Verify incident belongs to employee
	var incident models.Incident
	err = tx.Get(&incident, "SELECT * FROM incidents WHERE id = $1 AND employee_id = $2", incidentID, employeeID)
	if err != nil {
		return errors.New("incidente no encontrado o no autorizado")
	}

	// 2. Check if already has a justification
	var count int
	tx.Get(&count, "SELECT COUNT(*) FROM justifications WHERE incident_id = $1", incidentID)
	if count > 0 {
		return errors.New("este incidente ya tiene una justificación en proceso")
	}

	// 3. Create Justification
	now := time.Now()
	_, err = tx.Exec(`
		INSERT INTO justifications (incident_id, employee_id, message, evidence_url, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'pending', $5, $5)`,
		incidentID, employeeID, message, evidenceURL, now)
	
	if err != nil {
		return fmt.Errorf("error al crear justificación: %v", err)
	}

	// 4. Create System Alert for Admin
	alertMsg := fmt.Sprintf("Nueva justificación pendiente del empleado ID %d para el incidente #%d", employeeID, incidentID)
	if err := s.AlertService.CreateAlertTx(tx, "justification_submitted", "info", alertMsg, map[string]interface{}{
		"incident_id":    incidentID,
		"employee_id":    employeeID,
		"justification": message,
	}); err != nil {
		return fmt.Errorf("error creating alert: %v", err)
	}

	return tx.Commit()
}

func (s *JustificationService) ResolveJustification(id int, adminID int, approve bool, note string) error {
	var just models.Justification
	if err := s.DB.Get(&just, "SELECT * FROM justifications WHERE id = $1", id); err != nil {
		return errors.New("justificación no encontrada")
	}

	status := "rejected"
	incidentStatus := "pending"
	if approve {
		status = "approved"
		incidentStatus = "justified"
	}

	tx, err := s.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update Justification
	now := time.Now()
	_, err = tx.Exec(`
		UPDATE justifications 
		SET status = $1, resolved_by = $2, resolution_note = $3, updated_at = $4 
		WHERE id = $5`,
		status, adminID, note, now, id)
	if err != nil {
		return err
	}

	// Update Incident
	_, err = tx.Exec(`
		UPDATE incidents 
		SET status = $1, resolved_by = $2, resolution_note = $3, updated_at = $4 
		WHERE id = $5`,
		incidentStatus, adminID, note, now, just.IncidentID)
	if err != nil {
		return err
	}

	// If approved, we MUST recalculate the attendance earnings because deductions change
	if approve {
		var incident models.Incident
		tx.Get(&incident, "SELECT * FROM incidents WHERE id = $1", just.IncidentID)
		
		// This might be tricky if we don't have the service reference here...
		// But we can trigger a recalculation after the transaction.
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// Trigger Recalculation if approved
	if approve {
		var incident models.Incident
		s.DB.Get(&incident, "SELECT * FROM incidents WHERE id = $1", just.IncidentID)
		s.AttendanceService.RecalculateAttendance(s.DB, incident.AttendanceID)
	}

	return nil
}

func (s *JustificationService) ListPending(limit int) ([]models.Justification, error) {
	var list []models.Justification
	err := s.DB.Select(&list, "SELECT * FROM justifications WHERE status = 'pending' ORDER BY created_at ASC LIMIT $1", limit)
	return list, err
}
