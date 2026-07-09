package webutil

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

// ClientIP extracts the real client IP address from the request.
// It checks X-Forwarded-For and X-Real-IP headers first, then falls back to c.IP().
func ClientIP(c fiber.Ctx) string {
	if forwardedFor := c.Get("X-Forwarded-For"); forwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		return strings.TrimSpace(strings.Split(forwardedFor, ",")[0])
	}

	if realIP := c.Get("X-Real-IP"); realIP != "" {
		return strings.TrimSpace(realIP)
	}

	return c.IP()
}
