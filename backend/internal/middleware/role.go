package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func RoleCheck(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRoleSlug, ok := c.Locals("role_slug").(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: No role found in context"})
		}

		for _, role := range allowedRoles {
			if userRoleSlug == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden: Insufficient permissions for this action"})
	}
}
