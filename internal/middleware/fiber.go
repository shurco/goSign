package middleware

import (
	"time"

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

	"github.com/shurco/gosign/internal/config"
	"github.com/shurco/gosign/pkg/logging"
)

func requestLogger(logger *zerolog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		status := c.Response().StatusCode()

		// Escalate log level when the handler failed or produced a server error.
		event := logger.Info()
		if err != nil || status >= 500 {
			event = logger.Error().Err(err)
		}
		event.
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("url", c.OriginalURL()).
			Str("ip", c.IP()).
			Int("status", status).
			Dur("latency", time.Since(start)).
			Msg("http request")
		return err
	}
}

// Fiber applies common middleware to the Fiber app.
// With GOSIGN_CORS_ALLOWED_ORIGINS (or dev defaults when GOSIGN_DEV_MODE), credentialed
// cross-origin requests are allowed for those origins only. Otherwise credentials are off
// so Fiber v3 accepts the default wildcard AllowOrigins.
func Fiber(a *fiber.App, log *logging.Logger, cfg *config.Config) {
	corsCfg := cors.Config{
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
	}
	// Credentialed CORS requires an explicit allow-list; without one we fall back
	// to Fiber v3's default wildcard origin (no credentials).
	if len(cfg.CORSAllowedOrigins) > 0 {
		corsCfg.AllowCredentials = true
		corsCfg.AllowOrigins = cfg.CORSAllowedOrigins
	}

	a.Use(
		cors.New(corsCfg),
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

// rateLimitKey derives a stable rate-limit bucket from the authenticated
// principal (API key or user) and falls back to the client IP.
func rateLimitKey(c fiber.Ctx) string {
	auth := GetAuthContext(c)
	if auth != nil {
		switch auth.Type {
		case AuthTypeAPIKey:
			return "apikey:" + auth.UserID
		case AuthTypeJWT:
			return "user:" + auth.UserID
		}
	}
	return c.IP()
}

// RateLimiter creates rate limiting middleware with customizable limits.
func RateLimiter(max int, duration time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:          max,
		Expiration:   duration,
		KeyGenerator: rateLimitKey,
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
