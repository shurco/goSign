package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// GetUserID extracts user ID from request context
func GetUserID(c *fiber.Ctx) (string, error) {
	userID := c.Locals("user_id")
	if userID == nil {
		return "", fiber.NewError(fiber.StatusUnauthorized, "User not authenticated")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", fiber.NewError(fiber.StatusInternalServerError, "Invalid user context")
	}

	return userIDStr, nil
}

// GetOrganizationID extracts organization ID from request context
func GetOrganizationID(c *fiber.Ctx) (string, error) {
	orgID := c.Locals("organization_id")
	if orgID == nil {
		return "", nil // Organization ID is optional
	}

	orgIDStr, ok := orgID.(string)
	if !ok {
		return "", fiber.NewError(fiber.StatusInternalServerError, "Invalid organization context")
	}

	return orgIDStr, nil
}

// GetAccountID extracts account ID from request context or user's account
func GetAccountID(c *fiber.Ctx) (string, error) {
	// Try to get from context first
	accountID := c.Locals("account_id")
	if accountID != nil {
		accountIDStr, ok := accountID.(string)
		if ok && accountIDStr != "" {
			return accountIDStr, nil
		}
	}

	// Fallback: get from user's account_id via database
	// This requires UserQueries, so we'll handle it in the handler
	// For now, return error - middleware should set account_id
	return "", fiber.NewError(fiber.StatusUnauthorized, "Account not found in context")
}

// GetClientIP extracts the real client IP address from the request
// It checks X-Forwarded-For, X-Real-IP headers first, then falls back to c.IP()
func GetClientIP(c *fiber.Ctx) string {
	// Check X-Forwarded-For header (first IP in the chain)
	forwardedFor := c.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	realIP := c.Get("X-Real-IP")
	if realIP != "" {
		return strings.TrimSpace(realIP)
	}

	// Fallback to Fiber's IP() method
	return c.IP()
}

