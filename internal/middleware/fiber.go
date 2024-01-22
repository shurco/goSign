package middleware

import (
	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/earlydata"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/shurco/gosign/pkg/logging"
)

// Fiber is ...
func Fiber(a *fiber.App, log *logging.Logger) {
	a.Use(
		// cors.New(),
		earlydata.New(),
		helmet.New(),
		etag.New(),
		idempotency.New(),

		compress.New(compress.Config{
			Level: compress.LevelBestSpeed,
		}),

		fiberzerolog.New(fiberzerolog.Config{
			Logger: log.Logger,
		}),

		recover.New(),
	)
}
