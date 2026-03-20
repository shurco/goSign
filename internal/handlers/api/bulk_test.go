package api

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/testutil"
)

func TestBulkHandler_BulkCreateSubmissions(t *testing.T) {
	h := NewBulkHandler(nil)

	buildMultipart := func(templateID string, fileFieldName string, fileContent []byte) (*bytes.Buffer, string, error) {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		if templateID != "" {
			if err := writer.WriteField("template_id", templateID); err != nil {
				return nil, "", err
			}
		}
		if fileFieldName != "" {
			part, err := writer.CreateFormFile("file", fileFieldName)
			if err != nil {
				return nil, "", err
			}
			if _, err := part.Write(fileContent); err != nil {
				return nil, "", err
			}
		}
		if err := writer.Close(); err != nil {
			return nil, "", err
		}
		return &body, writer.FormDataContentType(), nil
	}

	tests := []struct {
		name         string
		setupApp     func() *fiber.App
		method       string
		path         string
		bodyFactory  func(t *testing.T) (*bytes.Buffer, string)
		expectedCode int
	}{
		{
			name: "no auth returns 401",
			setupApp: func() *fiber.App {
				app := fiber.New()
				app.Use(middleware.Protected())
				app.Post("/submissions/bulk", h.BulkCreateSubmissions)
				return app
			},
			method:       http.MethodPost,
			path:         "/submissions/bulk",
			bodyFactory:  func(t *testing.T) (*bytes.Buffer, string) { return bytes.NewBuffer(nil), "" },
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "missing template_id returns 400",
			setupApp: func() *fiber.App {
				app := fiber.New()
				app.Use(testutil.AuthMiddleware(testutil.User1))
				app.Post("/submissions/bulk", h.BulkCreateSubmissions)
				return app
			},
			method: http.MethodPost,
			path:   "/submissions/bulk",
			bodyFactory: func(t *testing.T) (*bytes.Buffer, string) {
				t.Helper()
				body, ct, err := buildMultipart("", "bulk.csv", []byte("title,email\nA,a@example.com\n"))
				if err != nil {
					t.Fatalf("build multipart: %v", err)
				}
				return body, ct
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "missing file returns 400",
			setupApp: func() *fiber.App {
				app := fiber.New()
				app.Use(testutil.AuthMiddleware(testutil.User1))
				app.Post("/submissions/bulk", h.BulkCreateSubmissions)
				return app
			},
			method: http.MethodPost,
			path:   "/submissions/bulk",
			bodyFactory: func(t *testing.T) (*bytes.Buffer, string) {
				t.Helper()
				body, ct, err := buildMultipart("tpl-1", "", nil)
				if err != nil {
					t.Fatalf("build multipart: %v", err)
				}
				return body, ct
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "invalid csv returns 400",
			setupApp: func() *fiber.App {
				app := fiber.New()
				app.Use(testutil.AuthMiddleware(testutil.User1))
				app.Post("/submissions/bulk", h.BulkCreateSubmissions)
				return app
			},
			method: http.MethodPost,
			path:   "/submissions/bulk",
			bodyFactory: func(t *testing.T) (*bytes.Buffer, string) {
				t.Helper()
				body, ct, err := buildMultipart("tpl-1", "broken.csv", []byte("title,email\n\"bad,a@example.com\n"))
				if err != nil {
					t.Fatalf("build multipart: %v", err)
				}
				return body, ct
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := tc.setupApp()
			body, contentType := tc.bodyFactory(t)

			req := httptest.NewRequest(tc.method, tc.path, body)
			if contentType != "" {
				req.Header.Set("Content-Type", contentType)
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
