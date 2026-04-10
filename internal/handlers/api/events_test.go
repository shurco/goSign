package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/testutil"
)

func TestEventHandler_List(t *testing.T) {
	pool := testutil.NewTestDB(t)
	h := NewEventHandler(pool)

	tests := []struct {
		name       string
		useAuth    bool
		limitQuery string
		wantLimit  int
		wantStatus int
	}{
		{
			name:       "no auth",
			useAuth:    false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "valid default limit",
			useAuth:    true,
			limitQuery: "",
			wantLimit:  10,
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid limit=5",
			useAuth:    true,
			limitQuery: "limit=5",
			wantLimit:  5,
			wantStatus: http.StatusOK,
		},
		{
			name:       "limit clipped to 100",
			useAuth:    true,
			limitQuery: "limit=200",
			wantLimit:  100,
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

			app.Get("/events", h.List)

			url := "/events"
			if tt.limitQuery != "" {
				url += "?" + tt.limitQuery
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			if tt.wantStatus != http.StatusOK {
				return
			}

			var body map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}

			data, ok := body["data"].(map[string]any)
			if !ok {
				t.Fatalf("body[\"data\"] is not map[string]any: %v", body["data"])
			}
			limitVal, ok := data["limit"].(float64)
			if !ok {
				t.Fatalf("data[\"limit\"] is not float64: %v", data["limit"])
			}
			got := int(limitVal)
			if got != tt.wantLimit {
				t.Fatalf("limit = %d, want %d", got, tt.wantLimit)
			}
		})
	}
}

