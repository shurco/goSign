package webutil

import (
	"github.com/gofiber/fiber/v2"
)

// HTTPResponse represents response body of API
type HTTPResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Response takes in context object, an HTTP status code, a message string and some data.
func Response(c *fiber.Ctx, code int, message string, data any) error {
	if len(message) > 0 {
		return c.Status(code).JSON(HTTPResponse{
			Success: code >= fiber.StatusOK && code < fiber.StatusMultipleChoices,
			Message: message,
			Data:    data,
		})
	}

	return c.Status(code).JSON(data)
}
