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

func TestInvitationHandler(t *testing.T) {
	pool := testutil.NewTestDB(t)
	orgQ := queries.NewOrganizationQueries(pool)
	userQ := queries.NewUserQueries(pool)

	invH := NewInvitationHandler(orgQ)
	memberH := NewMemberHandler(orgQ, userQ)

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

	// Create invitation via MemberHandler (so we get a valid token in DB).
	inviteApp := fiber.New()
	inviteApp.Use(testutil.AuthMiddleware(testutil.User1))
	inviteApp.Post("/organizations/:organization_id/members/invite", memberH.InviteMember)
	inviteReq := httptest.NewRequest(
		http.MethodPost,
		"/organizations/"+orgID+"/members/invite",
		bytes.NewReader([]byte(`{"email":"user2@gosign.local","role":"member"}`)),
	)
	inviteReq.Header.Set("Content-Type", "application/json")
	inviteResp, err := inviteApp.Test(inviteReq)
	if err != nil {
		t.Fatalf("invite app.Test: %v", err)
	}
	if inviteResp.StatusCode != http.StatusCreated {
		t.Fatalf("invite status = %d, want %d", inviteResp.StatusCode, http.StatusCreated)
	}

	invitations, err := orgQ.GetOrganizationInvitations(ctx, orgID)
	if err != nil {
		t.Fatalf("GetOrganizationInvitations: %v", err)
	}
	if len(invitations) == 0 {
		t.Fatalf("expected at least one invitation")
	}
	token := invitations[0].Token

	setupInvRoutes := func(app *fiber.App) {
		invH.RegisterRoutes(app.Group("/invitations"))
	}

	t.Run("AcceptInvitation no auth returns 401", func(t *testing.T) {
		app := fiber.New()
		setupInvRoutes(app)

		req := httptest.NewRequest(http.MethodPost, "/invitations/"+token+"/accept", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("AcceptInvitation empty token path returns 404 (route not matched)", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		setupInvRoutes(app)

		// Some routers capture empty segments as empty param; try the documented plan path style.
		req := httptest.NewRequest(http.MethodPost, "/invitations//accept", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})

	t.Run("AcceptInvitation non-existing token returns 404", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		setupInvRoutes(app)

		req := httptest.NewRequest(http.MethodPost, "/invitations/invalid-token/accept", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})

	t.Run("GetInvitationDetails non-existing token returns 404", func(t *testing.T) {
		app := fiber.New()
		setupInvRoutes(app)

		req := httptest.NewRequest(http.MethodGet, "/invitations/bad-token", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})

	t.Run("GetInvitationDetails valid token returns 200", func(t *testing.T) {
		app := fiber.New()
		// Even though this endpoint doesn't require user_id, keep auth to match the request pattern.
		app.Use(testutil.AuthMiddleware(testutil.User1))
		setupInvRoutes(app)

		req := httptest.NewRequest(http.MethodGet, "/invitations/"+token, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("AcceptInvitation invalid auth header still returns 401 via middleware.Protected", func(t *testing.T) {
		// This is a sanity check for auth-guard behavior in tests.
		app := fiber.New()
		app.Use(middleware.Protected())
		setupInvRoutes(app)

		req := httptest.NewRequest(http.MethodPost, "/invitations/"+token+"/accept", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})
}

