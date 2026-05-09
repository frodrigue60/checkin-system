package utils

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// SendError centraliza el manejo de errores para evitar fugas de información.
func SendError(c *fiber.Ctx, code int, message string, err error) error {
	if err != nil {
		GetLogger().Error(message, 
			zap.Error(err),
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
			zap.Int("status", code),
		)
	}
	
	return c.Status(code).JSON(fiber.Map{
		"error": message,
	})
}
