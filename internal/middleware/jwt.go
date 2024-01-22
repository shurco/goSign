package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

var mySigningKey = []byte("mysecretkey")

type MyCustomClaims struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// CreateToken is generate token with claims
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

// Parse With Claims
func ValidateToken(tokenString string) (*MyCustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	})
	if err != nil {
		return nil, errors.New("Unauthorize")
	}

	claims, ok := token.Claims.(*MyCustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("Unauthorize")
	}

	return claims, nil
}

// Protected is ...
func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		accessToken := c.Get("Authorization")
		accessToken = strings.Replace(accessToken, "Bearer ", "", 1)

		claims, err := ValidateToken(accessToken)
		if err != nil {
			return webutil.StatusUnauthorized(c, nil)
		}

		return webutil.StatusOK(c, "Claims", claims)
	}
}
