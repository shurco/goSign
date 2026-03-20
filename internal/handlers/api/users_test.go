package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/internal/testutil"
)

func TestUserHandler_GetCurrentUser(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userQ := queries.NewUserQueries(pool)
	h := NewUserHandler(userQ)

	tests := []struct {
		name       string
		setup      func(app *fiber.App)
		wantStatus int
		checkBody  func(t *testing.T, body map[string]any)
	}{
		{
			name: "no auth returns 401",
			setup: func(app *fiber.App) {
				app.Use(middleware.Protected())
				app.Get("/users/me", h.GetCurrentUser)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "valid user returns 200 with user data",
			setup: func(app *fiber.App) {
				app.Use(testutil.AuthMiddleware(testutil.User1))
				app.Get("/users/me", h.GetCurrentUser)
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]any) {
				data, ok := body["data"].(map[string]any)
				if !ok {
					t.Fatalf("body[\"data\"] is not map[string]any: %v", body["data"])
				}
				if got := data["email"]; got != testutil.User1.Email {
					t.Fatalf("email = %v, want %v", got, testutil.User1.Email)
				}
			},
		},
		{
			name: "unknown user ID returns 404",
			setup: func(app *fiber.App) {
				app.Use(testutil.AuthMiddleware(testutil.FixtureUser{
					ID: "00000000-0000-0000-0000-000000000000",
				}))
				app.Get("/users/me", h.GetCurrentUser)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			tt.setup(app)

			req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}

			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			if tt.checkBody != nil {
				var body map[string]any
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				tt.checkBody(t, body)
			}
		})
	}
}

