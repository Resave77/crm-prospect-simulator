package handler

import (
	"crm-prospect-prototype/service"

	"github.com/gofiber/fiber/v2"
)

func PlaceDetail(c *fiber.Ctx) error {

	placeID := c.Query("id")

	if placeID == "" {

		return c.Status(400).JSON(fiber.Map{
			"error": "Parameter id wajib diisi",
		})

	}

	result, err := service.PlaceDetail(placeID)

	if err != nil {

		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})

	}

	return c.Send(result)

}
