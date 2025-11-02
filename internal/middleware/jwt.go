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
	Id            string `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
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

// CreateToken generates JWT access token with claims (15 minutes)
func CreateToken(user *models.User) (string, error) {
	return CreateTokenWithOrg(user, "")
}

// CreateTokenWithOrg generates JWT access token with claims and organization ID (15 minutes)
func CreateTokenWithOrg(user *models.User, organizationID string) (string, error) {
	claims := MyCustomClaims{
		Id:             user.ID,
		Name:           user.Name,
		Email:          user.Email,
		OrganizationId: organizationID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(mySigningKey)
	return accessToken, err
}

// CreateRefreshToken generates JWT refresh token (7 days)
func CreateRefreshToken(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshToken, err := token.SignedString(mySigningKey)
	return refreshToken, err
}

// ValidateRefreshToken validates refresh token and returns user ID
func ValidateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return mySigningKey, nil
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
			return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", nil)
		}

		accessToken = strings.Replace(accessToken, "Bearer ", "", 1)
		claims, err := ValidateToken(accessToken)
		if err != nil {
			return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", nil)
		}

		// Store auth context
		c.Locals("auth", &AuthContext{
			Type:   AuthTypeJWT,
			UserID: claims.Id,
			Email:  claims.Email,
			Name:   claims.Name,
		})
		
		// Also store user_id and organization_id for easier access
		c.Locals("user_id", claims.Id)
		if claims.OrganizationId != "" {
			c.Locals("organization_id", claims.OrganizationId)
		}

		return c.Next()
	}
}

// RequireEmailVerification checks if user has verified their email
func RequireEmailVerification() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get auth context
		auth := GetAuthContext(c)
		if auth == nil {
			return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", nil)
		}

		// Only check for JWT auth (API keys don't need email verification)
		if auth.Type != AuthTypeJWT {
			return c.Next()
		}

		// Note: In a real implementation, you would check the database
		// to see if the user's email is verified. For now, we'll just
		// rely on the fact that the token was issued after verification.
		
		// TODO: Add database check for email verification status
		// user, err := queries.DB.GetUserByID(ctx, auth.UserID)
		// if err != nil || !user.EmailVerified {
		//     return webutil.StatusForbidden(c, "Email not verified")
		// }

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
