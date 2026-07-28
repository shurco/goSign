package routes

import (
	"github.com/gofiber/fiber/v3"

	"github.com/shurco/gosign/pkg/utils/webutil"
)

// NotFoundRoute func for describe 404 Error route.
func NotFoundRoute(a *fiber.App) {
	a.Use(func(c fiber.Ctx) error {
		return webutil.Response(c, fiber.StatusNotFound, "Not Found", nil)
	})
}
