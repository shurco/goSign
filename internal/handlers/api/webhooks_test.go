package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"

	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/testutil"
)

// memWebhookStore is an in-memory WebhookStore for handler tests.
type memWebhookStore struct {
	items  map[string]models.Webhook
	nextID int
}

func newMemWebhookStore() *memWebhookStore {
	return &memWebhookStore{items: make(map[string]models.Webhook)}
}

func (s *memWebhookStore) ListWebhooks(_ context.Context, accountID string) ([]models.Webhook, error) {
	var out []models.Webhook
	for _, w := range s.items {
		if w.AccountID == accountID {
			out = append(out, w)
		}
	}
	return out, nil
}

func (s *memWebhookStore) GetWebhook(_ context.Context, accountID, id string) (*models.Webhook, error) {
	w, ok := s.items[id]
	if !ok || w.AccountID != accountID {
		return nil, pgx.ErrNoRows
	}
	return &w, nil
}

func (s *memWebhookStore) CreateWebhook(_ context.Context, webhook *models.Webhook) error {
	s.nextID++
	webhook.ID = fmt.Sprintf("wh-%d", s.nextID)
	s.items[webhook.ID] = *webhook
	return nil
}

func (s *memWebhookStore) UpdateWebhook(_ context.Context, webhook *models.Webhook) error {
	existing, ok := s.items[webhook.ID]
	if !ok || existing.AccountID != webhook.AccountID {
		return pgx.ErrNoRows
	}
	s.items[webhook.ID] = *webhook
	return nil
}

func (s *memWebhookStore) DeleteWebhook(_ context.Context, accountID, id string) error {
	w, ok := s.items[id]
	if !ok || w.AccountID != accountID {
		return pgx.ErrNoRows
	}
	delete(s.items, id)
	return nil
}

func TestWebhookHandler_CRUD(t *testing.T) {
	store := newMemWebhookStore()
	h := NewWebhookHandler(store)

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
			path:       "/settings/webhooks/",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "List empty returns 200",
			useAuth:    true,
			method:     http.MethodGet,
			path:       "/settings/webhooks/",
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
			path:       "/settings/webhooks/",
			body:       []byte(`{"url":"https://x","events":["submission.created"],"secret":"s"}`),
			wantStatus: http.StatusCreated,
			check: func(t *testing.T, body map[string]any) {
				data := body["data"].(map[string]any)
				if data["account_id"] != testutil.User1.AccountID {
					t.Fatalf("account_id = %v, want %s", data["account_id"], testutil.User1.AccountID)
				}
				if data["enabled"] != true {
					t.Fatalf("enabled = %v, want true by default", data["enabled"])
				}
			},
		},
		{
			name:       "Create without events returns 400",
			useAuth:    true,
			method:     http.MethodPost,
			path:       "/settings/webhooks/",
			body:       []byte(`{"url":"https://x"}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Get not found returns 404",
			useAuth:    true,
			method:     http.MethodGet,
			path:       "/settings/webhooks/bad-id",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "Delete not found returns 404",
			useAuth:    true,
			method:     http.MethodDelete,
			path:       "/settings/webhooks/bad-id",
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

			h.RegisterRoutes(app.Group("/settings/webhooks"))

			var req *http.Request
			switch tt.method {
			case http.MethodGet, http.MethodDelete:
				req = httptest.NewRequest(tt.method, tt.path, nil)
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

func TestWebhookHandler_AccountIsolation(t *testing.T) {
	store := newMemWebhookStore()
	h := NewWebhookHandler(store)

	// Webhook belongs to User1's account.
	owned := &models.Webhook{
		AccountID: testutil.User1.AccountID,
		URL:       "https://x",
		Events:    []string{"submission.created"},
		Enabled:   true,
	}
	if err := store.CreateWebhook(context.Background(), owned); err != nil {
		t.Fatalf("seed webhook: %v", err)
	}

	app := fiber.New()
	app.Use(testutil.AuthMiddleware(testutil.User2))
	h.RegisterRoutes(app.Group("/settings/webhooks"))

	// User2 must not see or delete User1's webhook.
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/settings/webhooks/"+owned.ID, nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("get foreign webhook status = %d, want 404", resp.StatusCode)
	}

	resp, err = app.Test(httptest.NewRequest(http.MethodDelete, "/settings/webhooks/"+owned.ID, nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("delete foreign webhook status = %d, want 404", resp.StatusCode)
	}
}
