package handler

import (
	"encoding/json"

	"crm-prospect-prototype/service"

	"github.com/gofiber/fiber/v2"
)

func TextSearch(c *fiber.Ctx) error {

	body := map[string]string{

		"textQuery": "KFC Tebet",
	}

	request, _ := json.Marshal(body)

	result, err := service.TextSearch(request)

	if err != nil {

		return c.Status(500).JSON(fiber.Map{

			"error": err.Error(),
		})

	}

	return c.Send(result)

}
