package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

func Health(c *fiber.Ctx) error {
	return webutil.Response(c, fiber.StatusOK, "Pong", nil)
}
