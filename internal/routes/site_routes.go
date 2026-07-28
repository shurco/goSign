package routes

import (
	"github.com/gofiber/fiber/v3"

	handlers "github.com/shurco/gosign/internal/handlers/public"
)

// SiteRoutes wires public site endpoints.
func SiteRoutes(c *fiber.App, health *handlers.HealthHandler) {
	c.Get("/health", health.Health)
}
