package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/logging"
	"github.com/shurco/gosign/pkg/storage/redis"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// Validator interface for request validation
type Validator interface {
	Validate() error
}

// parseAndValidate parses request body and validates it
func parseAndValidate(c *fiber.Ctx, v Validator) error {
	if err := c.BodyParser(v); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	if err := v.Validate(); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	return nil
}

// createAuthTokens creates and stores access and refresh tokens
func createAuthTokens(ctx context.Context, user *queries.UserRecord) (access, refresh string, err error) {
	return createAuthTokensWithOrg(ctx, user, "")
}

// createAuthTokensWithOrg creates and stores access and refresh tokens with organization context
func createAuthTokensWithOrg(ctx context.Context, user *queries.UserRecord, organizationID string) (access, refresh string, err error) {
	modelUser := &models.User{
		ID:    user.ID,
		Name:  fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		Email: user.Email,
	}

	accessToken, err := middleware.CreateTokenWithOrg(modelUser, organizationID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := middleware.CreateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	// Store refresh token in Redis
	refreshKey := fmt.Sprintf("refresh_token:%s", refreshToken)
	if err := redis.Conn.Set(refreshKey, user.ID, 7*24*time.Hour); err != nil {
		logging.Log.Err(err).Msg("Failed to store refresh token")
	}

	return accessToken, refreshToken, nil
}

// invalidateRefreshToken removes refresh token from Redis
func invalidateRefreshToken(refreshToken string) {
	if refreshToken != "" {
		refreshKey := fmt.Sprintf("refresh_token:%s", refreshToken)
		if _, err := redis.Conn.Delete(refreshKey); err != nil {
			logging.Log.Err(err).Msg("Failed to delete refresh token")
		}
	}
}

