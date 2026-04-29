package handlers

import (
	"attendance-api/internal/models"
	"attendance-api/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"time"
)

type ManagerHandler struct {
	DB *sqlx.DB
}

// Get centers assigned to the current manager
func (h *ManagerHandler) ListManagedCenters(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	var centers []models.WorkCenter
	err := h.DB.Select(&centers, "SELECT * FROM work_centers WHERE manager_id = $1", userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}
	return c.JSON(centers)
}

// Get employees assigned to centers managed by the current user
func (h *ManagerHandler) ListManagedEmployees(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	var employees []struct {
		models.Employee
		UserName  string `db:"user_name" json:"user_name"`
		Email     string `db:"user_email" json:"user_email"`
		CenterName string `db:"center_name" json:"center_name"`
		ShiftName  *string `db:"shift_name" json:"shift_name"`
	}

	query := `
		SELECT e.*, u.name as user_name, u.email as user_email, wc.name as center_name, ws.name as shift_name
		FROM employees e
		JOIN users u ON e.user_id = u.id
		JOIN work_centers wc ON e.work_center_id = wc.id
		LEFT JOIN work_shifts ws ON e.work_shift_id = ws.id
		WHERE wc.manager_id = $1
	`
	err := h.DB.Select(&employees, query, userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}
	return c.JSON(employees)
}

// Assign employee to a new center or shift
func (h *ManagerHandler) AssignEmployee(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	employeeID := c.Params("id")

	var input struct {
		WorkCenterID *int `json:"work_center_id"`
		WorkShiftID  *int `json:"work_shift_id"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Security Check: Does the manager manage the employee's current center?
	// OR if they are assigning to a new center, do they manage that center?
	var currentCenterID int
	err := h.DB.Get(&currentCenterID, "SELECT work_center_id FROM employees WHERE id = $1", employeeID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Employee not found"})
	}

	// Verify manager oversees the current center
	var isManaged bool
	err = h.DB.Get(&isManaged, "SELECT EXISTS(SELECT 1 FROM work_centers WHERE id = $1 AND manager_id = $2)", currentCenterID, userID)
	if err != nil || !isManaged {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden: You do not manage this employee's current center"})
	}

	// If changing center, verify they manage the target center too
	if input.WorkCenterID != nil && *input.WorkCenterID != currentCenterID {
		var targetManaged bool
		h.DB.Get(&targetManaged, "SELECT EXISTS(SELECT 1 FROM work_centers WHERE id = $1 AND manager_id = $2)", *input.WorkCenterID, userID)
		if !targetManaged {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden: You do not manage the target center"})
		}
	}

	// Execute Update
	_, err = h.DB.Exec("UPDATE employees SET work_center_id = COALESCE($1, work_center_id), work_shift_id = $2, updated_at = $3 WHERE id = $4",
		input.WorkCenterID, input.WorkShiftID, time.Now(), employeeID)
	
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	return c.JSON(fiber.Map{"message": "Employee reassigned successfully"})
}




