package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// UserHandler handles user-related API requests
type UserHandler struct {
	userQueries *queries.UserQueries
}

// NewUserHandler creates a new user handler
func NewUserHandler(userQueries *queries.UserQueries) *UserHandler {
	return &UserHandler{
		userQueries: userQueries,
	}
}

// GetCurrentUser returns current authenticated user's information
// @Summary Get current user
// @Description Get current authenticated user's profile information
// @Tags users
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/users/me [get]
func (h *UserHandler) GetCurrentUser(c *fiber.Ctx) error {
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	// Get user from database
	user, err := h.userQueries.GetUserByID(c.Context(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to get user")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get user", nil)
	}

	// Return user data (without sensitive information like password)
	return webutil.Response(c, fiber.StatusOK, "User retrieved successfully", map[string]interface{}{
		"id":         user.ID,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"email":      user.Email,
		"role":       user.Role,
	})
}

// RegisterRoutes registers user routes
func (h *UserHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/me", h.GetCurrentUser)
}

