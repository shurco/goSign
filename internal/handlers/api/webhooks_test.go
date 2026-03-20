package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/testutil"
)

func TestWebhookHandler_CRUDViaResourceHandler(t *testing.T) {
	repo := newMemRepo[models.Webhook]()
	h := NewWebhookHandler(repo)

	tests := []struct {
		name       string
		useAuth    bool
		method     string
		path       string
		body       []byte
		wantStatus int
		check      func(t *testing.T, body map[string]any)
	}{
		{
			name:       "List no auth returns 401",
			useAuth:    false,
			method:     http.MethodGet,
			path:       "/webhooks/",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "List empty returns 200",
			useAuth:    true,
			method:     http.MethodGet,
			path:       "/webhooks/",
			wantStatus: http.StatusOK,
			check: func(t *testing.T, body map[string]any) {
				data := body["data"].(map[string]any)
				rawItems := data["items"]
				if rawItems == nil {
					return
				}
				items := rawItems.([]any)
				if len(items) != 0 {
					t.Fatalf("items length = %d, want 0", len(items))
				}
			},
		},
		{
			name:       "Create returns 201",
			useAuth:    true,
			method:     http.MethodPost,
			path:       "/webhooks/",
			body:       []byte(`{"url":"https://x","events":["submission.created"],"secret":"s","enabled":true}`),
			wantStatus: http.StatusCreated,
		},
		{
			name:       "Get not found returns 404",
			useAuth:    true,
			method:     http.MethodGet,
			path:       "/webhooks/bad-id",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			if tt.useAuth {
				app.Use(testutil.AuthMiddleware(testutil.User1))
			} else {
				app.Use(middleware.Protected())
			}

			h.RegisterRoutes(app.Group("/webhooks"))

			var req *http.Request
			switch tt.method {
			case http.MethodGet:
				req = httptest.NewRequest(http.MethodGet, tt.path, nil)
			case http.MethodPost:
				req = httptest.NewRequest(http.MethodPost, tt.path, bytes.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			default:
				t.Fatalf("unsupported method in test table: %s", tt.method)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			if tt.check == nil {
				return
			}

			var body map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			tt.check(t, body)
		})
	}
}

