package routes

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// NotFoundRoute func for describe 404 Error route.
func NotFoundRoute(a *fiber.App) {
	a.Use(func(c *fiber.Ctx) error {
		if strings.HasPrefix(c.Path(), "/api") {
			return webutil.StatusNotFound(c)
		}

		return c.Status(fiber.StatusNotFound).Render("404", nil, "layouts/clear")
	})
}
