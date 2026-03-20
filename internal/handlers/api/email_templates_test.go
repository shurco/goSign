package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/internal/testutil"
)

func TestEmailTemplateHandler(t *testing.T) {
	pool := testutil.NewTestDB(t)
	emailQ := &queries.EmailTemplateQueries{Pool: pool}
	userQ := queries.NewUserQueries(pool)
	h := NewEmailTemplateHandler(emailQ, userQ)

	routes := func(app *fiber.App) {
		app.Get("/email-templates", h.GetAllEmailTemplates)
		app.Get("/email-templates/:name", h.GetEmailTemplate)
		app.Post("/email-templates", h.CreateEmailTemplate)
		app.Delete("/email-templates/:id", h.DeleteEmailTemplate)
	}

	tplName := fmt.Sprintf("test-template-%d", time.Now().UnixNano())

	t.Run("GetAllEmailTemplates no auth returns 401", func(t *testing.T) {
		app := fiber.New()
		app.Use(middleware.Protected())
		routes(app)

		req := httptest.NewRequest(http.MethodGet, "/email-templates", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("GetAllEmailTemplates valid returns 200", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		routes(app)

		req := httptest.NewRequest(http.MethodGet, "/email-templates", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("CreateEmailTemplate missing fields returns 400", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		routes(app)

		req := httptest.NewRequest(http.MethodPost, "/email-templates", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}
	})

	t.Run("CreateEmailTemplate valid returns 201", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		routes(app)

		body := map[string]any{
			"name":    tplName,
			"content": "<p>hello</p>",
		}
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/email-templates", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
		}
	})

	t.Run("GetEmailTemplate not found returns 404", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		routes(app)

		req := httptest.NewRequest(http.MethodGet, "/email-templates/"+tplName+"-missing", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})

	t.Run("DeleteEmailTemplate not found returns 404", func(t *testing.T) {
		app := fiber.New()
		app.Use(testutil.AuthMiddleware(testutil.User1))
		routes(app)

		req := httptest.NewRequest(http.MethodDelete, "/email-templates/11111111-1111-1111-1111-111111111111", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})
}

