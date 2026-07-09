package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/shurco/gosign/internal/config"
	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

func signingKey() []byte {
	return []byte(config.Data().JWTSecret)
}

// AuthType defines the type of authentication used
type AuthType string

const (
	AuthTypeJWT    AuthType = "jwt"
	AuthTypeAPIKey AuthType = "api_key"
)

// AuthContext contains authentication information
type AuthContext struct {
	Type      AuthType
	UserID    string
	AccountID string
	Email     string
	Name      string
}

// MyCustomClaims represents JWT token claims
type MyCustomClaims struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	AccountId      string `json:"account_id,omitempty"`
	OrganizationId string `json:"organization_id,omitempty"`
	jwt.RegisteredClaims
}

// APIKeyValidator defines interface for API key validation
type APIKeyValidator interface {
	ValidateAPIKey(keyHash string) (*models.APIKey, error)
	UpdateLastUsed(keyID string) error
}

var apiKeyValidator APIKeyValidator

// SetAPIKeyValidator sets the validator for API keys
func SetAPIKeyValidator(validator APIKeyValidator) {
	apiKeyValidator = validator
}

// CreateTokenWithOrg generates JWT access token with claims and organization ID (15 minutes)
func CreateTokenWithOrg(user *models.User, organizationID string) (string, error) {
	claims := MyCustomClaims{
		Id:             user.ID,
		Name:           user.Name,
		Email:          user.Email,
		AccountId:      user.AccountID,
		OrganizationId: organizationID,
		RegisteredClaims: jwt.RegisteredClaims{
			// Keep user ID also in standard `jti` so callers/tests using RegisteredClaims.ID work.
			ID:        user.ID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signingKey())
}

// CreateRefreshToken generates JWT refresh token (7 days)
func CreateRefreshToken(userID string) (string, error) {
	// Include a unique ID so tokens differ even if issued within the same second.
	nonce := strconv.FormatInt(time.Now().UnixNano(), 10)
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ID:        nonce,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signingKey())
}

// ValidateRefreshToken validates refresh token and returns user ID
func ValidateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return signingKey(), nil
	})
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid refresh token")
	}

	return claims.Subject, nil
}

// ValidateToken parses and validates JWT token
func ValidateToken(tokenString string) (*MyCustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (any, error) {
		return signingKey(), nil
	})
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	claims, ok := token.Claims.(*MyCustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("unauthorized")
	}

	return claims, nil
}

// HashAPIKey creates SHA256 hash of API key
func HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// Protected authenticates requests using JWT or API Key
func Protected() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Try API Key first (X-API-Key header)
		apiKey := c.Get("X-API-Key")
		if apiKey != "" {
			if apiKeyValidator == nil {
				return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", nil)
			}

			keyHash := HashAPIKey(apiKey)
			keyModel, err := apiKeyValidator.ValidateAPIKey(keyHash)
			if err != nil {
				return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", nil)
			}

			// Check if key is enabled and not expired
			if !keyModel.Enabled {
				return webutil.Response(c, fiber.StatusForbidden, "API key is disabled", nil)
			}
			if keyModel.ExpiresAt != nil && keyModel.ExpiresAt.Before(time.Now()) {
				return webutil.Response(c, fiber.StatusForbidden, "API key has expired", nil)
			}

			go func(id string) {
				_ = apiKeyValidator.UpdateLastUsed(id)
			}(keyModel.ID)

			c.Locals("auth", &AuthContext{
				Type:      AuthTypeAPIKey,
				UserID:    keyModel.AccountID,
				AccountID: keyModel.AccountID,
			})
			c.Locals("user_id", keyModel.AccountID)
			c.Locals("account_id", keyModel.AccountID)

			return c.Next()
		}

		// Try JWT token (Authorization header)
		accessToken := c.Get("Authorization")
		if accessToken == "" {
			return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", nil)
		}

		accessToken = strings.TrimPrefix(accessToken, "Bearer ")
		claims, err := ValidateToken(accessToken)
		if err != nil {
			return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", nil)
		}

		// Store auth context
		c.Locals("auth", &AuthContext{
			Type:      AuthTypeJWT,
			UserID:    claims.Id,
			AccountID: claims.AccountId,
			Email:     claims.Email,
			Name:      claims.Name,
		})

		// Also store user_id, account_id and organization_id for easier access
		c.Locals("user_id", claims.Id)
		if claims.AccountId != "" {
			c.Locals("account_id", claims.AccountId)
		}
		if claims.OrganizationId != "" {
			c.Locals("organization_id", claims.OrganizationId)
		}

		return c.Next()
	}
}

// GetAuthContext retrieves authentication context from fiber locals
func GetAuthContext(c fiber.Ctx) *AuthContext {
	auth, ok := c.Locals("auth").(*AuthContext)
	if !ok {
		return nil
	}
	return auth
}
