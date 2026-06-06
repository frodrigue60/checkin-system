package handlers

import (
	"strconv"
	"time"

	"attendance-api/internal/models"
	"attendance-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func (h *HolidayHandler) ListHolidays(c *fiber.Ctx) error {
	var entities []models.Holiday
	if err := h.DB.Select(&entities, "SELECT * FROM holidays ORDER BY date"); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	dtos := make([]models.HolidayDTO, 0)
	for _, hol := range entities {
		dtos = append(dtos, models.MapHolidayToDTO(hol))
	}
	return c.JSON(dtos)
}

func (h *HolidayHandler) CreateHoliday(c *fiber.Ctx) error {
	var holiday models.Holiday
	if err := c.BodyParser(&holiday); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	now := time.Now()
	holiday.CreatedAt = &now
	holiday.UpdatedAt = &now

	query := `INSERT INTO holidays (name, date, description, type, created_at, updated_at) 
			  VALUES (:name, :date, :description, :type, :created_at, :updated_at) RETURNING id`
	
	rows, err := tx.NamedQuery(query, holiday)
	if err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}
	if rows.Next() {
		rows.Scan(&holiday.ID)
	}
	rows.Close()

	userID := c.Locals("user_id").(int)
	if err := h.AuditService.LogAction(c.Context(), tx, userID, models.AuditActionCreateHoliday, "holiday", holiday.ID, nil, holiday, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("holidays")
	return c.Status(fiber.StatusCreated).JSON(models.MapHolidayToDTO(holiday))
}

func (h *HolidayHandler) UpdateHoliday(c *fiber.Ctx) error {
	id := c.Params("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{Code: models.ErrInvalidID})
	}

	var holiday models.Holiday
	if err := c.BodyParser(&holiday); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	userID := c.Locals("user_id").(int)

	var old models.Holiday
	if err := tx.Get(&old, "SELECT * FROM holidays WHERE id = $1", idInt); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Holiday not found"})
	}

	_, err = tx.Exec("UPDATE holidays SET name = $1, date = $2, description = $3, type = $4, updated_at = $5 WHERE id = $6",
		holiday.Name, holiday.Date, holiday.Description, holiday.Type, time.Now(), idInt)
	
	if err != nil {
		tx.Rollback()
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}

	if err := h.AuditService.LogAction(c.Context(), tx, userID, models.AuditActionUpdateHoliday, "holiday", idInt, old, holiday, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("holidays")
	return c.JSON(fiber.Map{"message": "Holiday updated successfully"})
}

func (h *HolidayHandler) DeleteHoliday(c *fiber.Ctx) error {
	id := c.Params("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIError{Code: models.ErrInvalidID})
	}
	userID := c.Locals("user_id").(int)

	tx, err := h.DB.Beginx()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start transaction"})
	}

	var old models.Holiday
	if err := tx.Get(&old, "SELECT * FROM holidays WHERE id = $1", idInt); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Holiday not found"})
	}

	if _, err := tx.Exec("DELETE FROM holidays WHERE id = $1", idInt); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error deleting holiday"})
	}

	if err := h.AuditService.LogAction(c.Context(), tx, userID, models.AuditActionDeleteHoliday, "holiday", idInt, old, nil, c.IP()); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error logging action"})
	}

	if err := tx.Commit(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not commit transaction"})
	}

	h.Cache.Delete("holidays")
	return c.SendStatus(fiber.StatusNoContent)
}

// Helper to build attendance filters

