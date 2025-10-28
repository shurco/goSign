package webutil

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

// HTTPResponse represents response body of API
type HTTPResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Response is a takes in context object, an HTTP status code, a message string and some data.
func Response(c *fiber.Ctx, code int, message string, data any) error {
	if len(message) > 0 {
		return c.Status(code).JSON(HTTPResponse{
			Success: code == fiber.StatusOK,
			Message: message,
			Data:    data,
		})
	}

	return c.Status(code).JSON(data)
}

// StatusOK is ...
func StatusOK(c *fiber.Ctx, message string, data any) error {
	return Response(c, fiber.StatusOK, message, data)
}

// StatusUnauthorized is ...
func StatusUnauthorized(c *fiber.Ctx, data any) error {
	return Response(c, fiber.StatusUnauthorized, utils.StatusMessage(fiber.StatusUnauthorized), data)
}

// StatusNotFound is ...
func StatusNotFound(c *fiber.Ctx) error {
	return Response(c, fiber.StatusNotFound, utils.StatusMessage(fiber.StatusNotFound), nil)
}

// StatusNotFoundWithMessage is ...
func StatusNotFoundWithMessage(c *fiber.Ctx, message string) error {
	return Response(c, fiber.StatusNotFound, message, nil)
}

// StatusBadRequest is ...
func StatusBadRequest(c *fiber.Ctx, data any) error {
	return Response(c, fiber.StatusBadRequest, utils.StatusMessage(fiber.StatusBadRequest), data)
}

// StatusInternalServerError is ...
func StatusInternalServerError(c *fiber.Ctx) error {
	return Response(c, fiber.StatusInternalServerError, utils.StatusMessage(fiber.StatusInternalServerError), nil)
}

// StatusInternalServerErrorWithMessage is ...
func StatusInternalServerErrorWithMessage(c *fiber.Ctx, message string) error {
	return Response(c, fiber.StatusInternalServerError, message, nil)
}
