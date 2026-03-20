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

func TestSettingsHandler_ValidationAndAuth(t *testing.T) {
	h := NewSettingsHandler(nil, nil, nil, nil, nil)

	tests := []struct {
		name         string
		setupApp     func() *fiber.App
		method       string
		path         string
		body         string
		expectedCode int
	}{
		{
			name: "get no auth returns 401",
			setupApp: func() *fiber.App {
				app := fiber.New()
				app.Use(middleware.Protected())
				app.Get("/settings", h.Get)
				return app
			},
			method:       http.MethodGet,
			path:         "/settings",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "update email invalid json returns 400",
			setupApp: func() *fiber.App {
				app := fiber.New()
				app.Use(testutil.AuthMiddleware(testutil.User1))
				app.Put("/settings/email", h.UpdateEmail)
				return app
			},
			method:       http.MethodPut,
			path:         "/settings/email",
			body:         `{"provider":123}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "update email nil setting queries returns 500",
			setupApp: func() *fiber.App {
				app := fiber.New()
				app.Use(testutil.AuthMiddleware(testutil.User1))
				app.Put("/settings/email", h.UpdateEmail)
				return app
			},
			method:       http.MethodPut,
			path:         "/settings/email",
			body:         `{"provider":"smtp","smtp_host":"smtp.local","smtp_port":"1025","smtp_user":"u"}`,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "test sms missing to_phone returns 400",
			setupApp: func() *fiber.App {
				app := fiber.New()
				app.Use(testutil.AuthMiddleware(testutil.User1))
				app.Post("/settings/sms/test", h.TestSMS)
				return app
			},
			method:       http.MethodPost,
			path:         "/settings/sms/test",
			body:         `{}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "update branding without user queries returns 500",
			setupApp: func() *fiber.App {
				app := fiber.New()
				app.Use(testutil.AuthMiddleware(testutil.User1))
				app.Put("/settings/branding", h.UpdateBranding)
				return app
			},
			method:       http.MethodPut,
			path:         "/settings/branding",
			body:         `{"company_name":"Acme"}`,
			expectedCode: http.StatusInternalServerError,
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
