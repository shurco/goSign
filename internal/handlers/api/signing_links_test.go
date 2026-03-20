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

func TestSigningLinkHandler_AuthValidationAndSimpleDBFlow(t *testing.T) {
	pool := testutil.NewTestDB(t)
	hWithDB := NewSigningLinkHandler(pool, nil, nil)
	hNoDB := NewSigningLinkHandler(nil, nil, nil)

	tests := []struct {
		name         string
		setupApp     func() *fiber.App
		method       string
		path         string
		body         string
		expectedCode int
	}{
		{
			name: "create no auth returns 401",
			setupApp: func() *fiber.App {
				app := fiber.New()
				app.Use(middleware.Protected())
				app.Post("/signing-links", hNoDB.Create)
				return app
			},
			method:       http.MethodPost,
			path:         "/signing-links",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "create invalid json returns 500 (validator allows payload, template queries check fails)",
			setupApp: func() *fiber.App {
				app := fiber.New()
				app.Use(testutil.AuthMiddleware(testutil.User1))
				app.Post("/signing-links", hNoDB.Create)
				return app
			},
			method:       http.MethodPost,
			path:         "/signing-links",
			body:         `{"template_id":`,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "create nil template queries returns 500",
			setupApp: func() *fiber.App {
				app := fiber.New()
				app.Use(testutil.AuthMiddleware(testutil.User1))
				app.Post("/signing-links", hNoDB.Create)
				return app
			},
			method: http.MethodPost,
			path:   "/signing-links",
			body: `{
				"template_id":"tpl-1",
				"submitters":[{"name":"A","email":"a@example.com"}]
			}`,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "get detail with nil pool returns 500",
			setupApp: func() *fiber.App {
				app := fiber.New()
				app.Use(testutil.AuthMiddleware(testutil.User1))
				app.Get("/signing-links/:submission_id", hNoDB.Get)
				return app
			},
			method:       http.MethodGet,
			path:         "/signing-links/sub-1",
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "list with test db returns 200",
			setupApp: func() *fiber.App {
				app := fiber.New()
				app.Use(testutil.AuthMiddleware(testutil.User1))
				app.Get("/signing-links", hWithDB.List)
				return app
			},
			method:       http.MethodGet,
			path:         "/signing-links",
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := tc.setupApp()
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			if resp.StatusCode != tc.expectedCode {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tc.expectedCode)
			}
		})
	}
}
