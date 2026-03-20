package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/internal/testutil"
)

func TestMemberHandler_OrganizationMembersAndInvite(t *testing.T) {
	pool := testutil.NewTestDB(t)
	orgQ := queries.NewOrganizationQueries(pool)
	userQ := queries.NewUserQueries(pool)
	h := NewMemberHandler(orgQ, userQ)

	ctx := context.Background()

	orgID := uuid.New().String()
	org := &models.Organization{
		ID:      orgID,
		Name:    "Test Org",
		OwnerID: testutil.User1.AccountID,
	}
	if err := orgQ.CreateOrganization(ctx, org); err != nil {
		t.Fatalf("CreateOrganization: %v", err)
	}
	ownerMember := &models.OrganizationMember{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		UserID:         testutil.User1.AccountID,
		Role:           models.OrganizationRoleOwner,
	}
	if err := orgQ.AddOrganizationMember(ctx, ownerMember); err != nil {
		t.Fatalf("AddOrganizationMember: %v", err)
	}

	appRoute := func(app *fiber.App) {
		app.Get("/organizations/:organization_id/members", h.GetOrganizationMembers)
		app.Post("/organizations/:organization_id/members/invite", h.InviteMember)
	}

	t.Run("no auth GetOrganizationMembers returns 401", func(t *testing.T) {
		app := fiber.New()
		app.Use(middleware.Protected())
		appRoute(app)

		req := httptest.NewRequest(http.MethodGet, "/organizations/"+orgID+"/members", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("InviteMember missing email returns 400", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		appRoute(app)

		req := httptest.NewRequest(
			http.MethodPost,
			"/organizations/"+orgID+"/members/invite",
			bytes.NewReader([]byte(`{}`)),
		)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}
	})

	t.Run("InviteMember valid returns 201", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		appRoute(app)

		req := httptest.NewRequest(
			http.MethodPost,
			"/organizations/"+orgID+"/members/invite",
			bytes.NewReader([]byte(`{"email":"x@x.com","role":"member"}`)),
		)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
		}
	})
}

