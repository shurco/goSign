package testutil

import (
	"github.com/gofiber/fiber/v3"

	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/models"
)

// FixtureUser describes an authenticated fixture principal.
type FixtureUser struct {
	ID             string
	AccountID      string
	OrganizationID string
	Email          string
	Name           string
	Role           models.UserRole
}

var (
	AdminUser = FixtureUser{
		ID:        "ebf1ee29-ef5a-4aa9-8e7a-121fbcfc90bc",
		AccountID: "19ecfd4a-caf1-4ac9-91d7-21973fc9de31",
		Email:     "admin@gosign.local",
		Name:      "Admin User",
		Role:      models.UserRoleAdmin,
	}
	User1 = FixtureUser{
		ID:        "ef3a3b04-4d81-40a7-a387-cc572f68e23d",
		AccountID: "375507c0-2d39-4d80-915a-6e89522915a7",
		Email:     "user1@gosign.local",
		Name:      "User One",
		Role:      models.UserRoleUser,
	}
	User2 = FixtureUser{
		ID:        "b57349ba-8ce0-4606-a87b-c20a2848a0b2",
		AccountID: "c53aed39-0f8e-4926-843e-84db4a48de5c",
		Email:     "user2@gosign.local",
		Name:      "User Two",
		Role:      models.UserRoleUser,
	}
)

// AuthMiddleware injects auth locals for handler tests.
func AuthMiddleware(user FixtureUser) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Locals("auth", &middleware.AuthContext{
			Type:      middleware.AuthTypeJWT,
			UserID:    user.ID,
			AccountID: user.AccountID,
			Email:     user.Email,
			Name:      user.Name,
		})
		c.Locals("user_id", user.ID)
		c.Locals("account_id", user.AccountID)
		if user.OrganizationID != "" {
			c.Locals("organization_id", user.OrganizationID)
		}
		return c.Next()
	}
}
