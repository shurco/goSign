package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/testutil"
)

func TestBrandingHandler(t *testing.T) {
	h := NewBrandingHandler(nil)

	t.Run("GetBranding no auth returns 401", func(t *testing.T) {
		app := fiber.New()
		app.Use(middleware.Protected())
		app.Get("/settings/branding", h.GetBranding)

		req := httptest.NewRequest(http.MethodGet, "/settings/branding", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("GetBranding nil queries returns 500", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		app.Get("/settings/branding", h.GetBranding)

		req := httptest.NewRequest(http.MethodGet, "/settings/branding", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
		}
	})

	t.Run("UpdateBranding invalid body returns 400", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		app.Put("/settings/branding", h.UpdateBranding)

		req := httptest.NewRequest(http.MethodPut, "/settings/branding", bytes.NewReader([]byte(`{"branding":"bad"}`)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}
	})
}
