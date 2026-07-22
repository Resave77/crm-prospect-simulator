package response

import "github.com/gofiber/fiber/v2"

type ErrorBody struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"requestId,omitempty"`
}

func Data(c *fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(fiber.Map{
		"data": data,
		"meta": fiber.Map{"requestId": c.GetRespHeader(fiber.HeaderXRequestID)},
	})
}

func Error(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": ErrorBody{
		Code: code, Message: message, RequestID: c.GetRespHeader(fiber.HeaderXRequestID),
	}})
}
