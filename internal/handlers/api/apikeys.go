package api

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/services"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// APIKeyHandler handles API key operations
type APIKeyHandler struct {
	service *services.APIKeyService
}

// NewAPIKeyHandler creates new API key handler
func NewAPIKeyHandler(service *services.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{service: service}
}

// RegisterRoutes registers API key routes
func (h *APIKeyHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/", h.List)
	router.Post("/", h.Create)
	router.Patch("/:id", h.Update)
	router.Delete("/:id", h.Delete)
}

// CreateRequest represents API key creation request
type CreateRequest struct {
	Name      string `json:"name" validate:"required"`
	ExpiresAt *int64 `json:"expires_at,omitempty"` // Unix timestamp
}

// List lists all API keys for authenticated account
// @Summary List API keys
// @Description Get all API keys for the authenticated account
// @Tags apikeys
// @Accept json
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /v1/settings/api [get]
func (h *APIKeyHandler) List(c fiber.Ctx) error {
	auth := middleware.GetAuthContext(c)
	if auth == nil || auth.AccountID == "" {
		return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	keys, err := h.service.ListAccountKeys(auth.AccountID)
	if err != nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to list API keys", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "API keys retrieved successfully", keys)
}

// Create creates new API key
// @Summary Create API key
// @Description Create a new API key for the authenticated account
// @Tags apikeys
// @Accept json
// @Produce json
// @Param request body CreateRequest true "API key creation request"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /v1/settings/api [post]
func (h *APIKeyHandler) Create(c fiber.Ctx) error {
	auth := middleware.GetAuthContext(c)
	if auth == nil || auth.AccountID == "" {
		return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	var req CreateRequest
	if err := c.Bind().JSON(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "", nil)
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t := time.Unix(*req.ExpiresAt, 0)
		expiresAt = &t
	}

	// Generate new API key
	key, apiKey, err := h.service.CreateAPIKey(auth.AccountID, req.Name, expiresAt)
	if err != nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create API key", nil)
	}

	// Return plain key and masked model
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "API key created successfully. Save this key securely - it won't be shown again.",
		"data": fiber.Map{
			"key":     key, // Plain key - only shown once
			"api_key": apiKey,
		},
	})
}

// UpdateAPIKeyRequest represents API key update request
type UpdateAPIKeyRequest struct {
	Enabled *bool `json:"enabled" validate:"required"`
}

// Update updates API key state (enable/disable)
// @Summary Update API key
// @Description Enable or disable an API key
// @Tags apikeys
// @Accept json
// @Produce json
// @Param id path string true "API key ID"
// @Param body body UpdateAPIKeyRequest true "Update request"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /v1/settings/api/{id} [patch]
func (h *APIKeyHandler) Update(c fiber.Ctx) error {
	auth := middleware.GetAuthContext(c)
	if auth == nil {
		return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	var req UpdateAPIKeyRequest
	if err := c.Bind().JSON(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if req.Enabled == nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Field 'enabled' is required", nil)
	}

	keyID := c.Params("id")
	if *req.Enabled {
		if err := h.service.EnableKey(keyID); err != nil {
			return webutil.Response(c, fiber.StatusInternalServerError, "Failed to enable API key", nil)
		}
	} else {
		if err := h.service.DisableKey(keyID); err != nil {
			return webutil.Response(c, fiber.StatusInternalServerError, "Failed to disable API key", nil)
		}
	}

	return webutil.Response(c, fiber.StatusOK, "API key updated successfully", nil)
}

// Delete deletes API key
// @Summary Delete API key
// @Description Delete an API key permanently
// @Tags apikeys
// @Accept json
// @Produce json
// @Param id path string true "API key ID"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /v1/settings/api/{id} [delete]
func (h *APIKeyHandler) Delete(c fiber.Ctx) error {
	auth := middleware.GetAuthContext(c)
	if auth == nil {
		return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	keyID := c.Params("id")
	if err := h.service.DeleteKey(keyID); err != nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to delete API key", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "API key deleted successfully", nil)
}
