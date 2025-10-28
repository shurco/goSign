package middleware

import (
	"time"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/earlydata"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/shurco/gosign/pkg/logging"
)

// Fiber applies common middleware to the Fiber app
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

// RateLimiter creates rate limiting middleware with customizable limits
func RateLimiter(max int, duration time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: duration,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use API key or IP as rate limit key
			auth := GetAuthContext(c)
			if auth != nil && auth.Type == AuthTypeAPIKey {
				return "apikey:" + auth.UserID
			}
			if auth != nil && auth.Type == AuthTypeJWT {
				return "user:" + auth.UserID
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Rate limit exceeded. Please try again later.",
			})
		},
	})
}

// APIRateLimiter provides standard rate limiting for API endpoints
// Default: 100 requests per minute per user/API key
func APIRateLimiter() fiber.Handler {
	return RateLimiter(100, time.Minute)
}

// StrictRateLimiter provides stricter rate limiting for sensitive operations
// Default: 10 requests per minute per user/API key
func StrictRateLimiter() fiber.Handler {
	return RateLimiter(10, time.Minute)
}
