package handlers

import (
	"attendance-api/internal/services"
	"attendance-api/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type UploadHandler struct {
	Storage services.StorageService
}

type PresignedURLRequest struct {
	Files []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"files"`
}

type PresignedURLResponse struct {
	Items []PresignedURLItem `json:"items"`
}

type PresignedURLItem struct {
	FileName  string `json:"file_name"`
	UploadURL string `json:"upload_url"`
	PublicURL string `json:"public_url"`
	Key       string `json:"key"`
}

// UploadAttendanceEvidence returns presigned URLs for direct client-to-R2 upload
func (h *UploadHandler) UploadAttendanceEvidence(c *fiber.Ctx) error {
	if h.Storage == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Storage service is not configured",
		})
	}

	var req PresignedURLRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.Files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No files specified",
		})
	}

	// Extract userID from JWT claims (stored in locals by middleware)
	userID, ok := c.Locals("user_id").(int)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User ID not found in session"})
	}

	response := PresignedURLResponse{
		Items: make([]PresignedURLItem, 0, len(req.Files)),
	}

	for i, f := range req.Files {
		uploadURL, publicURL, key, err := h.Storage.GetPresignedUploadURL(c.Context(), userID, i+1, f.Name)
		if err != nil {
			return utils.SendError(c, fiber.StatusInternalServerError, "Failed to generate upload URL", err)
		}

		response.Items = append(response.Items, PresignedURLItem{
			FileName:  f.Name,
			UploadURL: uploadURL,
			PublicURL: publicURL,
			Key:       key,
		})
	}

	return c.JSON(response)
}
