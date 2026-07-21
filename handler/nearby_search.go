package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// NearbySearch serves JSON clients with the same dynamic category mapping as
// the CRM page instead of forcing every request to restaurant.
func NearbySearch(c *fiber.Ctx) error {
	latitude, latitudeErr := strconv.ParseFloat(strings.TrimSpace(c.Query("lat")), 64)
	longitude, longitudeErr := strconv.ParseFloat(strings.TrimSpace(c.Query("lng")), 64)
	radius, radiusErr := strconv.ParseFloat(strings.TrimSpace(c.Query("radius", "500")), 64)
	if latitudeErr != nil || longitudeErr != nil || radiusErr != nil || latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 || radius <= 0 || radius > 50000 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Lokasi atau radius pencarian tidak valid."})
	}

	categoryValue := strings.TrimSpace(c.Query("categories"))
	if categoryValue == "" {
		categoryValue = strings.TrimSpace(c.Query("category"))
	}
	if c.Method() != fiber.MethodGet {
		var payload struct {
			Category   string   `json:"category"`
			Categories []string `json:"categories"`
		}
		if c.BodyParser(&payload) == nil {
			if len(payload.Categories) > 0 {
				categoryValue = strings.Join(payload.Categories, ",")
			} else if strings.TrimSpace(payload.Category) != "" {
				categoryValue = payload.Category
			}
		}
	}

	selectedCategories := parseProspectCategories(categoryValue)
	if categoryValue != "" && len(selectedCategories) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Kategori pencarian tidak didukung."})
	}
	places, err := searchNearbyProspectCategories(latitude, longitude, radius, selectedCategories)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"success": false, "message": "Prospect belum dapat dimuat."})
	}
	return c.JSON(fiber.Map{"success": true, "places": places, "count": len(places)})
}
