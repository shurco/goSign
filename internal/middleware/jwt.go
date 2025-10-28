package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

var mySigningKey = []byte("mysecretkey")

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
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
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

// CreateToken generates JWT token with claims
func CreateToken(user *models.User) (string, error) {
	claims := MyCustomClaims{
		user.ID,
		user.Name,
		user.Email,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(mySigningKey)
	return accessToken, err
}

// ValidateToken parses and validates JWT token
func ValidateToken(tokenString string) (*MyCustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (any, error) {
		return mySigningKey, nil
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
	return func(c *fiber.Ctx) error {
		// Try API Key first (X-API-Key header)
		apiKey := c.Get("X-API-Key")
		if apiKey != "" {
			if apiKeyValidator == nil {
				return webutil.StatusUnauthorized(c, nil)
			}

			keyHash := HashAPIKey(apiKey)
			keyModel, err := apiKeyValidator.ValidateAPIKey(keyHash)
			if err != nil {
				return webutil.StatusUnauthorized(c, nil)
			}

			// Check if key is enabled and not expired
			if !keyModel.Enabled {
				return webutil.Response(c, fiber.StatusForbidden, "API key is disabled", nil)
			}
			if keyModel.ExpiresAt != nil && keyModel.ExpiresAt.Before(time.Now()) {
				return webutil.Response(c, fiber.StatusForbidden, "API key has expired", nil)
			}

			// Update last used timestamp (async, don't block request)
			go apiKeyValidator.UpdateLastUsed(keyModel.ID)

			// Store auth context
			c.Locals("auth", &AuthContext{
				Type:      AuthTypeAPIKey,
				UserID:    keyModel.ID,
				AccountID: keyModel.AccountID,
			})

			return c.Next()
		}

		// Try JWT token (Authorization header)
		accessToken := c.Get("Authorization")
		if accessToken == "" {
			return webutil.StatusUnauthorized(c, nil)
		}

		accessToken = strings.Replace(accessToken, "Bearer ", "", 1)
		claims, err := ValidateToken(accessToken)
		if err != nil {
			return webutil.StatusUnauthorized(c, nil)
		}

		// Store auth context
		c.Locals("auth", &AuthContext{
			Type:   AuthTypeJWT,
			UserID: claims.Id,
			Email:  claims.Email,
			Name:   claims.Name,
		})

		return c.Next()
	}
}

// GetAuthContext retrieves authentication context from fiber locals
func GetAuthContext(c *fiber.Ctx) *AuthContext {
	auth, ok := c.Locals("auth").(*AuthContext)
	if !ok {
		return nil
	}
	return auth
}
