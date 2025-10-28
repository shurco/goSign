# Swagger API Documentation

**Last Updated**: 2025-10-27 20:00 UTC

## 📚 Overview

This guide explains how to generate, view, and use Swagger/OpenAPI documentation for the goSign API. The interactive Swagger UI provides a complete reference for all API endpoints with the ability to test them directly.

## 🚀 Quick Start

### Install swag

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### Generate Documentation

```bash
make swagger
# or
swag init -g cmd/goSign/main.go -o cmd/goSign/docs --parseDependency --parseInternal
```

### View Documentation

After starting the application, documentation is available at:

- **Swagger UI**: http://localhost:8088/swagger/index.html
- **JSON Specification**: http://localhost:8088/swagger/doc.json
- **YAML Specification**: http://localhost:8088/swagger/doc.yaml

## 🔐 Authentication in Swagger UI

### JWT Bearer Token

1. Sign in via `/auth/signin` endpoint
2. Copy the received token
3. Click "Authorize" button in Swagger UI
4. Enter token in format: `Bearer YOUR_TOKEN`
5. Click "Authorize" and "Close"

### API Key

1. Create an API key via Settings or `/api/v1/apikeys` endpoint
2. Click "Authorize" button in Swagger UI
3. Select "ApiKeyAuth"
4. Paste your API key (without any prefix)
5. Click "Authorize" and "Close"

## 📝 Annotation Structure

### General Info (cmd/goSign/docs.go)

```go
// @title           goSign API
// @version         2.0
// @description     Document signing platform with multi-signer workflows
// @host            localhost:8088
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
```

### Endpoint Annotations (in handlers)

```go
// @Summary      Brief description
// @Description  Detailed description of what the endpoint does
// @Tags         tag-name
// @Accept       json
// @Produce      json
// @Param        id path string true "Resource ID"
// @Param        body body RequestType true "Request body"
// @Success      200 {object} ResponseType
// @Failure      400 {object} map[string]any "Bad request"
// @Failure      401 {object} map[string]any "Unauthorized"
// @Failure      404 {object} map[string]any "Not found"
// @Security     BearerAuth
// @Router       /endpoint/{id} [get]
func Handler(c *fiber.Ctx) error {
    // implementation
}
```

## 🔧 Parameter Types

- `path` - URL path parameter (e.g., `/users/{id}`)
- `query` - Query string parameter (e.g., `?page=1`)
- `header` - HTTP header
- `body` - Request body (JSON)
- `formData` - Form data (multipart/form-data)

## 📖 Usage Examples

### GET Request with Query Parameters

```go
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Page size" default(20)
// @Param status query string false "Filter by status" Enums(draft, pending, completed)
```

### POST Request with Body

```go
// @Param body body CreateSubmissionRequest true "Submission data"
// @Success 201 {object} models.Submission
```

### File Upload

```go
// @Param file formData file true "PDF file to upload"
// @Accept multipart/form-data
```

### Authentication

```go
// @Security BearerAuth
// or
// @Security ApiKeyAuth
```

## ✅ Best Practices

1. **Always specify response type**: `{object} ModelName` or `{array} ModelName`
2. **Use enums** for limited value sets
3. **Document all error codes**: 400, 401, 404, 500
4. **Group endpoints** by tags (Submissions, Templates, etc.)
5. **Add examples** in descriptions for complex requests
6. **Keep descriptions clear** and concise
7. **Use consistent naming** across endpoints

## 🔄 Updating Documentation

After changing annotations:

```bash
make swagger
```

Swagger UI will update automatically when you restart the application.

## 🐛 Troubleshooting

### swag: command not found

```bash
export PATH=$PATH:$(go env GOPATH)/bin
# Add to ~/.bashrc or ~/.zshrc
```

### Documentation Not Updating

1. Delete old docs: `rm -rf cmd/goSign/docs/`
2. Regenerate: `make swagger`
3. Restart application

### Models Not Displaying

Ensure that:
- You use `--parseDependency` flag
- Models are exported (capitalized names)
- JSON tags are specified correctly
- Models are imported in handlers

### Incorrect Type Generation

If types are generated incorrectly:
- Check that struct fields have proper `json` tags
- Ensure nested structs are also exported
- Verify that `map[string]interface{}` is changed to `map[string]any`

## 📍 API Endpoints Reference

### Authentication
- `POST /auth/signin` - User sign in
- `POST /auth/signout` - User sign out

### Submissions (Protected)
- `GET /api/v1/submissions` - List submissions
- `GET /api/v1/submissions/:id` - Get submission details
- `POST /api/v1/submissions` - Create submission
- `PUT /api/v1/submissions/:id` - Update submission
- `DELETE /api/v1/submissions/:id` - Delete submission
- `POST /api/v1/submissions/send` - Send invitations to signers
- `POST /api/v1/submissions/bulk` - Bulk create from CSV
- `POST /api/v1/submissions/expire` - Mark as expired

### Submitters (Protected)
- `GET /api/v1/submitters` - List submitters
- `GET /api/v1/submitters/:id` - Get submitter details
- `POST /api/v1/submitters` - Create submitter
- `PUT /api/v1/submitters/:id` - Update submitter
- `DELETE /api/v1/submitters/:id` - Delete submitter
- `POST /api/v1/submitters/resend` - Resend invitation
- `POST /api/v1/submitters/decline` - Decline signing
- `POST /api/v1/submitters/complete` - Complete signing

### Templates (Protected)
- `GET /api/v1/templates` - List templates
- `GET /api/v1/templates/:id` - Get template details
- `POST /api/v1/templates` - Create template
- `PUT /api/v1/templates/:id` - Update template
- `DELETE /api/v1/templates/:id` - Delete template
- `POST /api/v1/templates/clone` - Clone template
- `POST /api/v1/templates/from-file` - Create from PDF file

### Webhooks (Protected)
- `GET /api/v1/webhooks` - List webhooks
- `GET /api/v1/webhooks/:id` - Get webhook details
- `POST /api/v1/webhooks` - Create webhook
- `PUT /api/v1/webhooks/:id` - Update webhook
- `DELETE /api/v1/webhooks/:id` - Delete webhook

### API Keys (Protected)
- `GET /api/v1/apikeys` - List API keys
- `GET /api/v1/apikeys/:id` - Get API key details
- `POST /api/v1/apikeys` - Create API key
- `PUT /api/v1/apikeys/:id` - Update API key
- `DELETE /api/v1/apikeys/:id` - Delete/revoke API key
- `POST /api/v1/apikeys/:id/enable` - Enable API key
- `POST /api/v1/apikeys/:id/disable` - Disable API key

### Settings (Protected)
- `GET /api/v1/settings` - Get all settings
- `PUT /api/v1/settings/email` - Update email settings
- `PUT /api/v1/settings/storage` - Update storage settings
- `PUT /api/v1/settings/branding` - Update branding

### Public Endpoints
- `POST /verify/pdf` - Verify PDF signature
- `POST /sign` - Sign PDF document
- `GET /s/:slug` - Submitter signing page
- `POST /s/:slug` - Save signature
- `GET /embed/sign` - Get embedded signing URL

## 📚 Additional Resources

- [API Authentication Guide](API_AUTHENTICATION.md)
- [Embedded Signing Guide](EMBEDDED_SIGNING.md)
- [Implementation Details](IMPLEMENTATION_COMPLETE.md)
- [Swagger Official Docs](https://github.com/swaggo/swag)

---

**Status**: ✅ Complete  
**Version**: 2.0.0  
**Last Updated**: 2025-10-27 20:00 UTC
