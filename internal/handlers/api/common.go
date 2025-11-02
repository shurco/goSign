package api

import (
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

