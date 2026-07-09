package services

import (
	"fmt"
	"time"

	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/storage/redis"
)

// IssueAuthTokens creates an access/refresh token pair for the user and stores
// the refresh token in Redis. organizationID is embedded into the access token
// claims when non-empty.
func IssueAuthTokens(user *queries.UserRecord, organizationID string) (access, refresh string, err error) {
	modelUser := &models.User{
		ID:        user.ID,
		AccountID: user.AccountID,
		Name:      fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		Email:     user.Email,
	}

	access, err = middleware.CreateTokenWithOrg(modelUser, organizationID)
	if err != nil {
		return "", "", fmt.Errorf("failed to create access token: %w", err)
	}

	refresh, err = middleware.CreateRefreshToken(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	// The refresh flow requires this key in Redis; treat store failure as fatal.
	if err := redis.Conn.Set(fmt.Sprintf("refresh_token:%s", refresh), user.ID, 7*24*time.Hour); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return access, refresh, nil
}
