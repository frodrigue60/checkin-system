package handlers

import (
	"fmt"

	"attendance-api/internal/models"
	"attendance-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func (h *AdminHandler) ListAuditLogs(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	if page < 1 { page = 1 }
	if limit < 1 { limit = 20 }
	offset := (page - 1) * limit

	action := c.Query("action")
	entity := c.Query("entity")
	start := c.Query("start")
	end := c.Query("end")

	where := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if (action != "" && action != "all") {
		where += fmt.Sprintf(" AND l.action ILIKE $%d", argCount)
		args = append(args, "%"+action+"%")
		argCount++
	}
	if (entity != "" && entity != "all") {
		where += fmt.Sprintf(" AND l.entity_type ILIKE $%d", argCount)
		args = append(args, "%"+entity+"%")
		argCount++
	}
	if start != "" {
		where += fmt.Sprintf(" AND l.created_at >= $%d", argCount)
		args = append(args, start)
		argCount++
	}
	if end != "" {
		where += fmt.Sprintf(" AND l.created_at <= $%d", argCount)
		args = append(args, end)
		argCount++
	}

	var total int
	h.DB.Get(&total, "SELECT COUNT(*) FROM audit_logs l " + where, args...)

	logs := []models.AuditLog{}
	query := fmt.Sprintf(`
		SELECT l.*, COALESCE(u.name, 'System') as user_name
		FROM audit_logs l
		LEFT JOIN users u ON l.user_id = u.id
		%s
		ORDER BY l.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argCount, argCount+1)
	
	args = append(args, limit, offset)

	if err := h.DB.Select(&logs, query, args...); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	totalPages := (total + limit - 1) / limit

	return c.JSON(fiber.Map{
		"data":        logs,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// SYSTEM ALERTS
func (h *AdminHandler) ListAlerts(c *fiber.Ctx) error {
	onlyUnread := c.Query("unread") == "true"
	limit := c.QueryInt("limit", 50)

	alerts, err := h.AlertService.ListAlerts(c.Context(), onlyUnread, limit)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	return c.JSON(alerts)
}

func (h *AdminHandler) MarkAlertAsRead(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	if err := h.AlertService.MarkAsRead(c.Context(), id); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// JUSTIFICATIONS
func (h *AdminHandler) ListJustifications(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)
	list, err := h.JustificationService.ListPending(c.Context(), h.DB, limit)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}
	return c.JSON(list)
}

func (h *AdminHandler) ResolveJustification(c *fiber.Ctx) error {
	id, _ := c.ParamsInt("id")
	var req struct {
		Approve bool   `json:"approve"`
		Note    string `json:"note"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	adminID := c.Locals("user_id").(int)
	if err := h.JustificationService.ResolveJustification(c.Context(), h.DB, id, adminID, req.Approve, req.Note); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Justificaci├│n resuelta correctamente"})
}

// BULK ACTIONS

