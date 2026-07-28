# Swagger API Documentation

**Last Updated**: 2026-07-28

## Overview

goSign handlers are annotated with [swag](https://github.com/swaggo/swag) comments (`@Summary`, `@Router`, ...). The general API info lives in `cmd/server/docs.go`. The application does **not** serve a Swagger UI itself — you generate an OpenAPI specification from the annotations and view it with any external tool.

## Generating the Specification

### Install swag

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### Generate

```bash
swag init -g cmd/server/main.go -o /tmp/gosign-swagger --parseDependency --parseInternal
```

This produces `swagger.json` / `swagger.yaml` in the output directory. View it with any OpenAPI viewer, for example:

```bash
# Swagger UI in Docker
docker run --rm -p 8090:8080 \
  -e SWAGGER_JSON=/spec/swagger.json \
  -v /tmp/gosign-swagger:/spec swaggerapi/swagger-ui
```

The generated directory is a build artifact — do not commit it.

## Annotation Structure

### General Info (cmd/server/docs.go)

```go
// @title goSign API
// @description API for goSign digital document signing platform
// @host localhost:8088
// @BasePath /v1
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
func Handler(c fiber.Ctx) error {
    // implementation
}
```

## Parameter Types

- `path` - URL path parameter (e.g., `/users/{id}`)
- `query` - Query string parameter (e.g., `?page=1`)
- `header` - HTTP header
- `body` - Request body (JSON)
- `formData` - Form data (multipart/form-data)

## Usage Examples

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

## Best Practices

1. **Always specify response type**: `{object} ModelName` or `{array} ModelName`
2. **Use enums** for limited value sets
3. **Document all error codes**: 400, 401, 404, 500
4. **Group endpoints** by tags (Submissions, Templates, etc.)
5. **Keep `@Router` paths in sync** with `RegisterRoutes` registrations
6. **Use consistent naming** across endpoints

## Troubleshooting

### swag: command not found

```bash
export PATH=$PATH:$(go env GOPATH)/bin
# Add to ~/.bashrc or ~/.zshrc
```

### Models Not Displaying

Ensure that:
- You use `--parseDependency` flag
- Models are exported (capitalized names)
- JSON tags are specified correctly

## Route Reference

The authoritative list of routes is in the code:
- `internal/routes/api_routes.go` — route groups and middleware
- `internal/handlers/api/*.go` — `RegisterRoutes` methods of each handler
- `internal/handlers/public/*.go` — public auth, verification, signing, and embed endpoints

## Additional Resources

- [API Authentication Guide](API_AUTHENTICATION.md)
- [Embedded Signing Guide](EMBEDDED_SIGNING.md)
- [Swagger Official Docs](https://github.com/swaggo/swag)
