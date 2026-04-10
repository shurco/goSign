package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/internal/testutil"
)

func TestOrganizationHandler_CRUD(t *testing.T) {
	pool := testutil.NewTestDB(t)
	orgQ := queries.NewOrganizationQueries(pool)
	userQ := queries.NewUserQueries(pool)
	h := NewOrganizationHandler(orgQ, userQ)

	nonExistingOrgID := "11111111-1111-1111-1111-111111111111"

	routes := func(app *fiber.App) {
		app.Post("/organizations", h.CreateOrganization)
		app.Get("/organizations", h.GetUserOrganizations)
		app.Get("/organizations/:organization_id", h.GetOrganization)
		app.Put("/organizations/:organization_id", h.UpdateOrganization)
		app.Delete("/organizations/:organization_id", h.DeleteOrganization)
	}

	t.Run("CreateOrganization no auth returns 401", func(t *testing.T) {
		app := fiber.New()
		app.Use(middleware.Protected())
		routes(app)

		req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("CreateOrganization missing name returns 400", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		routes(app)

		req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}
	})

	t.Run("CreateOrganization valid then GetOrganization returns 200", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		routes(app)

		req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewReader([]byte(`{"name":"Test Org"}`)))
		req.Header.Set("Content-Type", "application/json")
		createResp, err := app.Test(req)
		if err != nil {
			t.Fatalf("create app.Test: %v", err)
		}
		if createResp.StatusCode != http.StatusCreated {
			t.Fatalf("create status = %d, want %d", createResp.StatusCode, http.StatusCreated)
		}

		var created map[string]any
		if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
			t.Fatalf("decode create response: %v", err)
		}
		data, ok := created["data"].(map[string]any)
		if !ok {
			t.Fatalf("created[\"data\"] is not map[string]any: %v", created["data"])
		}
		org, ok := data["organization"].(map[string]any)
		if !ok {
			t.Fatalf("data[\"organization\"] is not map[string]any: %v", data["organization"])
		}
		orgID, ok := org["id"].(string)
		if !ok {
			t.Fatalf("org[\"id\"] is not string: %v", org["id"])
		}

		getReq := httptest.NewRequest(http.MethodGet, "/organizations/"+orgID, nil)
		getResp, err := app.Test(getReq)
		if err != nil {
			t.Fatalf("get app.Test: %v", err)
		}
		if getResp.StatusCode != http.StatusOK {
			t.Fatalf("get status = %d, want %d", getResp.StatusCode, http.StatusOK)
		}
	})

	t.Run("GetUserOrganizations no auth returns 401", func(t *testing.T) {
		app := fiber.New()
		app.Use(middleware.Protected())
		routes(app)

		req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("GetUserOrganizations valid returns 200", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		routes(app)

		req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("GetOrganization not found returns 404", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		routes(app)

		req := httptest.NewRequest(http.MethodGet, "/organizations/"+nonExistingOrgID, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})

	t.Run("UpdateOrganization not found returns 403", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		routes(app)

		req := httptest.NewRequest(http.MethodPut, "/organizations/"+nonExistingOrgID, bytes.NewReader([]byte(`{"name":"x"}`)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusForbidden)
		}
	})

	t.Run("DeleteOrganization not found returns 403", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		routes(app)

		req := httptest.NewRequest(http.MethodDelete, "/organizations/"+nonExistingOrgID, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusForbidden)
		}
	})
}

