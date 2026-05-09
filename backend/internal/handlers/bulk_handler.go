package handlers

import (
	"fmt"

	"attendance-api/internal/models"
	"attendance-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func (h *BulkHandler) BulkUpdateEmployees(c *fiber.Ctx) error {
	var req models.BulkEmployeeUpdate
	if err := utils.ParseAndValidate(c, &req); err != nil {
		return err
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Could not start transaction", err)
	}

	userID := c.Locals("user_id").(int)
	ip := c.IP()

	for _, id := range req.IDs {
		var old models.Employee
		if err := tx.Get(&old, "SELECT * FROM employees WHERE id = $1", id); err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusNotFound, fmt.Sprintf("Employee ID %d not found", id), err)
		}

		query := "UPDATE employees SET updated_at = NOW()"
		args := map[string]interface{}{"id": id}

		if req.WorkCenterID != nil {
			query += ", work_center_id = :wc"
			args["wc"] = *req.WorkCenterID
		}
		if req.WorkShiftID != nil {
			query += ", work_shift_id = :ws"
			args["ws"] = *req.WorkShiftID
		}
		if req.IsActive != nil {
			query += ", is_active = :active"
			args["active"] = *req.IsActive
		}
		query += " WHERE id = :id"

		if _, err := tx.NamedExec(query, args); err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update employee batch", err)
		}

		if err := h.AuditService.LogAction(c.Context(), tx, userID, models.AuditActionBulkUpdateEmployees, "employee", id, old, args, ip); err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Error logging batch action", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to commit transaction", err)
	}

	h.Cache.Delete("employees") // Invalidate employee cache if it exists
	return c.JSON(fiber.Map{"message": fmt.Sprintf("Successfully updated %d employees", len(req.IDs))})
}

func (h *BulkHandler) BulkDeleteEmployees(c *fiber.Ctx) error {
	var req models.BulkRequest
	if err := utils.ParseAndValidate(c, &req); err != nil {
		return err
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Could not start transaction", err)
	}

	userID := c.Locals("user_id").(int)
	ip := c.IP()

	for _, id := range req.IDs {
		var old models.Employee
		if err := tx.Get(&old, "SELECT * FROM employees WHERE id = $1", id); err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusNotFound, fmt.Sprintf("Employee ID %d not found", id), err)
		}

		if _, err := tx.Exec("DELETE FROM employees WHERE id = $1", id); err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusBadRequest, fmt.Sprintf("Cannot delete employee %d: likely has linked attendance records", id), err)
		}

		if err := h.AuditService.LogAction(c.Context(), tx, userID, models.AuditActionBulkDeleteEmployees, "employee", id, old, nil, ip); err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Error logging batch action", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to commit transaction", err)
	}

	h.Cache.Delete("employees")
	return c.JSON(fiber.Map{"message": fmt.Sprintf("Successfully deleted %d employees", len(req.IDs))})
}

func (h *BulkHandler) BulkJustifyAttendances(c *fiber.Ctx) error {
	var req models.BulkJustifyRequest
	if err := utils.ParseAndValidate(c, &req); err != nil {
		return err
	}

	adminID := c.Locals("user_id").(int)
	tx, err := h.DB.Beginx()
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Could not start transaction", err)
	}

	for _, id := range req.AttendanceIDs {
		if _, err := tx.Exec("UPDATE incidents SET status = $1, resolution_note = $2, resolved_by = $3 WHERE attendance_id = $4 AND status = $5", 
			models.StatusJustified, req.Note, adminID, id, models.StatusPending); err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update incidents", err)
		}

		if err := h.AttendanceService.AutoDetectIncidents(c.Context(), tx, id); err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to recalculate totals", err)
		}

		h.AuditService.LogAction(c.Context(), tx, adminID, models.AuditActionBulkJustifyAttendances, "attendance", id, nil, req.Note, c.IP())
	}

	if err := tx.Commit(); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to commit transaction", err)
	}

	return c.JSON(fiber.Map{"message": fmt.Sprintf("Successfully justified %d attendance records", len(req.AttendanceIDs))})
}

// INCIDENTS

func (h *BulkHandler) BulkJustifyIncidents(c *fiber.Ctx) error {
	var req models.BulkJustifyIncidentsRequest
	if err := utils.ParseAndValidate(c, &req); err != nil {
		return err
	}

	adminID := c.Locals("user_id").(int)
	tx, err := h.DB.Beginx()
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Could not start transaction", err)
	}

	// Track affected attendances for recalculation
	attendanceIDs := make(map[int]bool)

	for _, id := range req.IncidentIDs {
		var attID int
		err := tx.Get(&attID, "UPDATE incidents SET status = $1, resolution_note = $2, resolved_by = $3 WHERE id = $4 AND status = $5 RETURNING attendance_id",
			models.StatusJustified, req.Note, adminID, id, models.StatusPending)
		
		if err != nil {
			// Skip if not found or already resolved, but in a real scenario we might want to log this
			continue
		}
		attendanceIDs[attID] = true
	}

	for attID := range attendanceIDs {
		if err := h.AttendanceService.AutoDetectIncidents(c.Context(), tx, attID); err != nil {
			tx.Rollback()
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to recalculate attendance totals", err)
		}
		h.AuditService.LogAction(c.Context(), tx, adminID, models.AuditActionBulkJustifyAttendances, "attendance", attID, nil, "Bulk Incident Justification: "+req.Note, c.IP())
	}

	if err := tx.Commit(); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to commit transaction", err)
	}

	return c.JSON(fiber.Map{"message": fmt.Sprintf("Successfully justified %d incidents", len(req.IncidentIDs))})
}






