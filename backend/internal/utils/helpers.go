package utils

import (
	"attendance-api/internal/models"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// ParseID extracts and validates an integer ID from a route parameter.
// Returns the parsed ID or sends a 400 error response with the INVALID_ID code.
func ParseID(c *fiber.Ctx, paramName string) (int, error) {
	raw := c.Params(paramName)
	id, err := strconv.Atoi(raw)
	if err != nil {
		return 0, c.Status(fiber.StatusBadRequest).JSON(models.APIError{
			Code:    models.ErrInvalidID,
			Message: fmt.Sprintf("Invalid %s format", paramName),
		})
	}
	return id, nil
}

// SafeLocalsInt extracts an integer value from fiber.Ctx Locals with defensive type checking.
// Returns 0 and an error response if the value is missing or not an int.
func SafeLocalsInt(c *fiber.Ctx, key string) (int, bool) {
	val := c.Locals(key)
	if val == nil {
		return 0, false
	}
	intVal, ok := val.(int)
	return intVal, ok
}
