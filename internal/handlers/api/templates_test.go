package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/testutil"
)

func TestTemplateHandler_ValidationAndAuth(t *testing.T) {
	h := NewTemplateHandler(newMemRepo[models.Template](), nil)

	tests := []struct {
		name         string
		setupApp     func() *fiber.App
		method       string
		path         string
		body         string
		expectedCode int
	}{
		{
			name: "search no auth returns 401",
			setupApp: func() *fiber.App {
				app := fiber.New()
				app.Use(middleware.Protected())
				app.Get("/templates/search", h.SearchTemplates)
				return app
			},
			method:       http.MethodGet,
			path:         "/templates/search",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "create empty invalid json returns 400",
			setupApp: func() *fiber.App {
				app := fiber.New()
				app.Use(testutil.AuthMiddleware(testutil.User1))
				app.Post("/templates/empty", h.CreateEmptyTemplate)
				return app
			},
			method:       http.MethodPost,
			path:         "/templates/empty",
			body:         `{"name":123}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "create empty missing name returns 400",
			setupApp: func() *fiber.App {
				app := fiber.New()
				app.Use(testutil.AuthMiddleware(testutil.User1))
				app.Post("/templates/empty", h.CreateEmptyTemplate)
				return app
			},
			method:       http.MethodPost,
			path:         "/templates/empty",
			body:         `{}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "create from file missing file returns 400",
			setupApp: func() *fiber.App {
				app := fiber.New()
				app.Use(testutil.AuthMiddleware(testutil.User1))
				app.Post("/templates/from-file", h.CreateFromType)
				return app
			},
			method:       http.MethodPost,
			path:         "/templates/from-file",
			body:         `{"name":"Doc","type":"pdf"}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "add to favorites invalid json returns 400",
			setupApp: func() *fiber.App {
				app := fiber.New()
				app.Use(testutil.AuthMiddleware(testutil.User1))
				app.Post("/templates/favorites", h.AddToFavorites)
				return app
			},
			method:       http.MethodPost,
			path:         "/templates/favorites",
			body:         `{"template_id":123}`,
			expectedCode: http.StatusBadRequest,
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
