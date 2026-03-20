package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/earlydata"
	"github.com/gofiber/fiber/v3/middleware/etag"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/idempotency"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/rs/zerolog"
	"time"

	"github.com/shurco/gosign/pkg/logging"
)

func requestLogger(logger *zerolog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		status := c.Response().StatusCode()

		evt := logger.Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("url", c.OriginalURL()).
			Str("ip", c.IP()).
			Int("status", status).
			Dur("latency", time.Since(start))

		// If the handler returned an error or we produced 5xx - bump log level.
		if err != nil || status >= 500 {
			evt = logger.Error().
				Str("method", c.Method()).
				Str("path", c.Path()).
				Str("url", c.OriginalURL()).
				Str("ip", c.IP()).
				Int("status", status).
				Dur("latency", time.Since(start)).
				Err(err)
		}

		evt.Msg("http request")
		return err
	}
}

// Fiber applies common middleware to the Fiber app
func Fiber(a *fiber.App, log *logging.Logger) {
	a.Use(
		cors.New(cors.Config{
			AllowCredentials: true,
			AllowHeaders: []string{
				"Origin",
				"Content-Type",
				"Accept",
				"Authorization",
				"X-API-Key",
				"X-Organization-ID",
			},
			AllowMethods: []string{
				"GET",
				"POST",
				"PUT",
				"PATCH",
				"DELETE",
				"OPTIONS",
			},
		}),
		earlydata.New(),
		helmet.New(),
		etag.New(),
		idempotency.New(),

		compress.New(compress.Config{
			Level: compress.LevelBestSpeed,
		}),

		requestLogger(log.Logger),

		recover.New(),
	)
}

// RateLimiter creates rate limiting middleware with customizable limits
func RateLimiter(max int, duration time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: duration,
		KeyGenerator: func(c fiber.Ctx) string {
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
		LimitReached: func(c fiber.Ctx) error {
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
