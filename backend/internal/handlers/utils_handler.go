package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"attendance-api/internal/utils"
	"github.com/bradfitz/latlong"
	"github.com/gofiber/fiber/v2"
)

type UtilsHandler struct{}

type ParseMapsRequest struct {
	URL string `json:"url"`
}

// ExtractCoordsFromURL extracts latitude and longitude from a Google Maps URL
// It looks for patterns like @lat,lng or !3dlat!4dlng
func ExtractCoordsFromURL(url string) (string, string, error) {
	// Pattern 1: @-34.5824982,-58.4116962
	reAt := regexp.MustCompile(`@(-?\d+\.\d+),(-?\d+\.\d+)`)
	matchesAt := reAt.FindStringSubmatch(url)
	if len(matchesAt) >= 3 {
		return matchesAt[1], matchesAt[2], nil
	}

	// Pattern 2: !3d-34.5824982!4d-58.4116962
	reDetails := regexp.MustCompile(`!3d(-?\d+\.\d+)!4d(-?\d+\.\d+)`)
	matchesDetails := reDetails.FindStringSubmatch(url)
	if len(matchesDetails) >= 3 {
		return matchesDetails[1], matchesDetails[2], nil
	}

	return "", "", fmt.Errorf("coordinates not found in URL")
}

// @Summary Parse Google Maps URL to extract coordinates
// @Tags Utils
// @Accept json
// @Produce json
// @Param request body ParseMapsRequest true "Maps URL"
// @Success 200 {object} map[string]string "Success: {latitude: '...', longitude: '...'}"
// @Failure 400 {object} map[string]string "Error: Invalid request"
// @Router /api/v1/admin/utils/parse-maps-url [post]
func (h *UtilsHandler) ParseMapsURL(c *fiber.Ctx) error {
	var req ParseMapsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// We only want the last URL, we don't need to actually download the page
			return nil
		},
	}

	// Try to resolve the URL (short URLs redirect to long ones)
	resp, err := client.Get(req.URL)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error", err)
	}
	defer resp.Body.Close()

	finalURL := resp.Request.URL.String()
	lat, lng, err := ExtractCoordsFromURL(finalURL)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"latitude":  lat,
		"longitude": lng,
	})
}

// @Summary Detect timezone from coordinates
// @Tags Utils
// @Produce json
// @Param lat query number true "Latitude"
// @Param lng query number true "Longitude"
// @Success 200 {object} map[string]string "Success: {timezone: '...'}"
// @Router /api/v1/admin/utils/detect-timezone [get]
func (h *UtilsHandler) DetectTimezone(c *fiber.Ctx) error {
	lat := c.Query("lat")
	lng := c.Query("lng")

	if lat == "" || lng == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Latitude and longitude are required"})
	}

	fLat, errLat := strconv.ParseFloat(lat, 64)
	fLng, errLng := strconv.ParseFloat(lng, 64)

	if errLat != nil || errLng != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid coordinates"})
	}

	zone := latlong.LookupZoneName(fLat, fLng)
	if zone == "" {
		zone = "UTC"
	}

	return c.JSON(fiber.Map{"timezone": zone})
}





