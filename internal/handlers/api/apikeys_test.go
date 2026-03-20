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
	"github.com/shurco/gosign/internal/services"
	"github.com/shurco/gosign/internal/testutil"
)

func TestAPIKeyHandler(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := queries.NewAPIKeyRepository(pool)
	service := services.NewAPIKeyService(repo)
	h := NewAPIKeyHandler(service)

	routes := func(app *fiber.App) {
		app.Get("/apikeys", h.List)
		app.Post("/apikeys", h.Create)
		app.Put("/apikeys/:id/enable", h.Enable)
		app.Put("/apikeys/:id/disable", h.Disable)
		app.Delete("/apikeys/:id", h.Delete)
	}

	t.Run("list no auth returns 401", func(t *testing.T) {
		app := fiber.New()
		app.Use(middleware.Protected())
		routes(app)

		req := httptest.NewRequest(http.MethodGet, "/apikeys", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("list without account_id returns 401", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.FixtureUser{
			ID: testutil.User1.ID,
		}))
		routes(app)

		req := httptest.NewRequest(http.MethodGet, "/apikeys", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("list valid user returns 200", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		routes(app)

		req := httptest.NewRequest(http.MethodGet, "/apikeys", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("create invalid json returns 400", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		routes(app)

		// name must be string; send number to force JSON binding error.
		body := []byte(`{"name":123}`)
		req := httptest.NewRequest(http.MethodPost, "/apikeys", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}
	})

	t.Run("create then enable disable delete", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		routes(app)

		createBody := []byte(`{"name":"mykey"}`)
		createReq := httptest.NewRequest(http.MethodPost, "/apikeys", bytes.NewReader(createBody))
		createReq.Header.Set("Content-Type", "application/json")

		createResp, err := app.Test(createReq)
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
		plainKey, ok := data["key"].(string)
		if !ok {
			t.Fatalf("data[\"key\"] is not string: %v", data["key"])
		}
		apiKeyMap, ok := data["api_key"].(map[string]any)
		if !ok {
			t.Fatalf("data[\"api_key\"] is not map[string]any: %v", data["api_key"])
		}
		apiKeyID, ok := apiKeyMap["id"].(string)
		if !ok {
			t.Fatalf("api_key[\"id\"] is not string: %v", apiKeyMap["id"])
		}
		// Enable/Disable routes use key_hash as :id; Delete uses the UUID.
		keyHash := middleware.HashAPIKey(plainKey)

		enableResp, err := app.Test(httptest.NewRequest(http.MethodPut, "/apikeys/"+keyHash+"/enable", nil))
		if err != nil {
			t.Fatalf("enable app.Test: %v", err)
		}
		if enableResp.StatusCode != http.StatusOK {
			t.Fatalf("enable status = %d, want %d", enableResp.StatusCode, http.StatusOK)
		}

		disableResp, err := app.Test(httptest.NewRequest(http.MethodPut, "/apikeys/"+keyHash+"/disable", nil))
		if err != nil {
			t.Fatalf("disable app.Test: %v", err)
		}
		if disableResp.StatusCode != http.StatusOK {
			t.Fatalf("disable status = %d, want %d", disableResp.StatusCode, http.StatusOK)
		}

		deleteResp, err := app.Test(httptest.NewRequest(http.MethodDelete, "/apikeys/"+apiKeyID, nil))
		if err != nil {
			t.Fatalf("delete app.Test: %v", err)
		}
		if deleteResp.StatusCode != http.StatusOK {
			t.Fatalf("delete status = %d, want %d", deleteResp.StatusCode, http.StatusOK)
		}
	})
}
