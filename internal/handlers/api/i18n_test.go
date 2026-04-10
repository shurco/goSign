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

func TestI18nHandler_GetLocales(t *testing.T) {
	h := NewI18nHandler(nil, nil)

	tests := []struct {
		name       string
		useAuth    bool
		wantStatus int
		check      func(t *testing.T, body map[string]any)
	}{
		{
			name:       "no auth returns 401",
			useAuth:    false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 200 with locales",
			useAuth:    true,
			wantStatus: http.StatusOK,
			check: func(t *testing.T, body map[string]any) {
				data := body["data"].(map[string]any)
				uiLocales := data["ui_locales"].(map[string]any)
				if got, want := len(uiLocales), 7; got != want {
					t.Fatalf("ui_locales count = %d, want %d", got, want)
				}
			},
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
			app.Get("/i18n/locales", h.GetLocales)

			req := httptest.NewRequest(http.MethodGet, "/i18n/locales", nil)
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

func TestI18nHandler_UpdateUserLocale(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userQ := queries.NewUserQueries(pool)
	accountQ := queries.NewAccountQueries(pool)
	h := NewI18nHandler(userQ, accountQ)

	tests := []struct {
		name       string
		useAuth    bool
		body       any
		wantStatus int
	}{
		{
			name:       "no auth",
			useAuth:    false,
			body:       nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "valid locale",
			useAuth:    true,
			body:       map[string]any{"locale": "ru"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid locale",
			useAuth:    true,
			body:       map[string]any{"locale": "xx"},
			wantStatus: http.StatusBadRequest,
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
			app.Put("/i18n/user/locale", h.UpdateUserLocale)

			var reqBody *bytes.Reader
			if tt.body != nil {
				b, err := json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("marshal request: %v", err)
				}
				reqBody = bytes.NewReader(b)
			} else {
				reqBody = bytes.NewReader(nil)
			}

			req := httptest.NewRequest(http.MethodPut, "/i18n/user/locale", reqBody)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestI18nHandler_UpdateAccountLocale(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userQ := queries.NewUserQueries(pool)
	accountQ := queries.NewAccountQueries(pool)
	h := NewI18nHandler(userQ, accountQ)

	tests := []struct {
		name       string
		useAuth    bool
		body       any
		wantStatus int
	}{
		{
			name:       "no auth",
			useAuth:    false,
			body:       nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "valid locale",
			useAuth:    true,
			body:       map[string]any{"locale": "de"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid locale",
			useAuth:    true,
			body:       map[string]any{"locale": "zz"},
			wantStatus: http.StatusBadRequest,
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
			app.Put("/i18n/account/locale", h.UpdateAccountLocale)

			var reqBody *bytes.Reader
			if tt.body != nil {
				b, err := json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("marshal request: %v", err)
				}
				reqBody = bytes.NewReader(b)
			} else {
				reqBody = bytes.NewReader(nil)
			}

			req := httptest.NewRequest(http.MethodPut, "/i18n/account/locale", reqBody)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

