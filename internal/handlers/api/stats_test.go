package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/testutil"
)

func TestStatsHandler_Get(t *testing.T) {
	pool := testutil.NewTestDB(t)
	h := NewStatsHandler(pool)

	tests := []struct {
		name       string
		useAuth    bool
		wantStatus int
	}{
		{
			name:       "no auth",
			useAuth:    false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "valid user",
			useAuth:    true,
			wantStatus: http.StatusOK,
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

			app.Get("/stats", h.Get)
			req := httptest.NewRequest(http.MethodGet, "/stats", nil)
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

