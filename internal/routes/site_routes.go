package routes

import (
	"github.com/gofiber/fiber/v3"

	handlers "github.com/shurco/gosign/internal/handlers/public"
)

// SiteRoutes is ...
func SiteRoutes(c *fiber.App) {
	c.Get("/health", handlers.Health)
}
