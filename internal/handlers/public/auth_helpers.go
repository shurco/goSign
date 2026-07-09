package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/shurco/gosign/pkg/logging"
	"github.com/shurco/gosign/pkg/storage/redis"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// parseAndValidate parses request body and validates it
func parseAndValidate(c fiber.Ctx, v any) error {
	if err := c.Bind().JSON(v); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	if err := webutil.ValidateStruct(v); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	return nil
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
