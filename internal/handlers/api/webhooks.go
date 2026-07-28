package api

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// WebhookStore is the persistence contract used by WebhookHandler.
// Implemented by *queries.WebhookQueries.
type WebhookStore interface {
	ListWebhooks(ctx context.Context, accountID string) ([]models.Webhook, error)
	GetWebhook(ctx context.Context, accountID, id string) (*models.Webhook, error)
	CreateWebhook(ctx context.Context, webhook *models.Webhook) error
	UpdateWebhook(ctx context.Context, webhook *models.Webhook) error
	DeleteWebhook(ctx context.Context, accountID, id string) error
}

// WebhookHandler handles account-scoped webhook CRUD
type WebhookHandler struct {
	store WebhookStore
}

// NewWebhookHandler creates new handler
func NewWebhookHandler(store WebhookStore) *WebhookHandler {
	return &WebhookHandler{store: store}
}

// webhookRequest is the create/update payload.
type webhookRequest struct {
	URL     string   `json:"url"`
	Events  []string `json:"events"`
	Secret  string   `json:"secret"`
	Enabled *bool    `json:"enabled"`
}

func (r *webhookRequest) validate() string {
	if r.URL == "" {
		return "url is required"
	}
	if len(r.Events) == 0 {
		return "at least one event is required"
	}
	return ""
}

// enabledOrDefault returns the requested enabled state, defaulting to true.
func (r *webhookRequest) enabledOrDefault() bool {
	return r.Enabled == nil || *r.Enabled
}

// List returns all webhooks of the authenticated account
// @Summary List webhooks
// @Description Returns all webhooks configured for the authenticated account
// @Tags webhooks
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /v1/settings/webhooks [get]
func (h *WebhookHandler) List(c fiber.Ctx) error {
	auth := middleware.GetAuthContext(c)
	if auth == nil || auth.AccountID == "" {
		return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	webhooks, err := h.store.ListWebhooks(c.Context(), auth.AccountID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list webhooks")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to retrieve webhooks", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "webhooks", map[string]any{
		"items": webhooks,
		"total": len(webhooks),
	})
}

// Get returns a webhook by ID
// @Summary Get webhook
// @Description Returns a single webhook by ID
// @Tags webhooks
// @Produce json
// @Param id path string true "Webhook ID"
// @Success 200 {object} models.Webhook
// @Failure 404 {object} map[string]any
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /v1/settings/webhooks/{id} [get]
func (h *WebhookHandler) Get(c fiber.Ctx) error {
	auth := middleware.GetAuthContext(c)
	if auth == nil || auth.AccountID == "" {
		return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	webhook, err := h.store.GetWebhook(c.Context(), auth.AccountID, c.Params("id"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return webutil.Response(c, fiber.StatusNotFound, "Webhook not found", nil)
		}
		log.Error().Err(err).Msg("Failed to get webhook")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to retrieve webhook", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "webhook", webhook)
}

// Create creates a new webhook
// @Summary Create webhook
// @Description Creates a webhook subscription for the authenticated account
// @Tags webhooks
// @Accept json
// @Produce json
// @Param body body webhookRequest true "Webhook data"
// @Success 201 {object} models.Webhook
// @Failure 400 {object} map[string]any
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /v1/settings/webhooks [post]
func (h *WebhookHandler) Create(c fiber.Ctx) error {
	auth := middleware.GetAuthContext(c)
	if auth == nil || auth.AccountID == "" {
		return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	var req webhookRequest
	if err := c.Bind().JSON(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if msg := req.validate(); msg != "" {
		return webutil.Response(c, fiber.StatusBadRequest, msg, nil)
	}

	webhook := &models.Webhook{
		AccountID: auth.AccountID,
		URL:       req.URL,
		Events:    req.Events,
		Secret:    req.Secret,
		Enabled:   req.enabledOrDefault(),
	}
	if err := h.store.CreateWebhook(c.Context(), webhook); err != nil {
		log.Error().Err(err).Msg("Failed to create webhook")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create webhook", nil)
	}

	return webutil.Response(c, fiber.StatusCreated, "webhook", webhook)
}

// Update updates an existing webhook
// @Summary Update webhook
// @Description Updates a webhook by ID
// @Tags webhooks
// @Accept json
// @Produce json
// @Param id path string true "Webhook ID"
// @Param body body webhookRequest true "Webhook data"
// @Success 200 {object} models.Webhook
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /v1/settings/webhooks/{id} [put]
func (h *WebhookHandler) Update(c fiber.Ctx) error {
	auth := middleware.GetAuthContext(c)
	if auth == nil || auth.AccountID == "" {
		return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	var req webhookRequest
	if err := c.Bind().JSON(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if msg := req.validate(); msg != "" {
		return webutil.Response(c, fiber.StatusBadRequest, msg, nil)
	}

	webhook := &models.Webhook{
		ID:        c.Params("id"),
		AccountID: auth.AccountID,
		URL:       req.URL,
		Events:    req.Events,
		Secret:    req.Secret,
		Enabled:   req.enabledOrDefault(),
	}
	if err := h.store.UpdateWebhook(c.Context(), webhook); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return webutil.Response(c, fiber.StatusNotFound, "Webhook not found", nil)
		}
		log.Error().Err(err).Msg("Failed to update webhook")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to update webhook", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "webhook", webhook)
}

// Delete removes a webhook
// @Summary Delete webhook
// @Description Deletes a webhook by ID
// @Tags webhooks
// @Param id path string true "Webhook ID"
// @Success 204 "No content"
// @Failure 404 {object} map[string]any
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /v1/settings/webhooks/{id} [delete]
func (h *WebhookHandler) Delete(c fiber.Ctx) error {
	auth := middleware.GetAuthContext(c)
	if auth == nil || auth.AccountID == "" {
		return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	if err := h.store.DeleteWebhook(c.Context(), auth.AccountID, c.Params("id")); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return webutil.Response(c, fiber.StatusNotFound, "Webhook not found", nil)
		}
		log.Error().Err(err).Msg("Failed to delete webhook")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to delete webhook", nil)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// RegisterRoutes registers all routes for webhooks
func (h *WebhookHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/", h.List)
	router.Get("/:id", h.Get)
	router.Post("/", h.Create)
	router.Put("/:id", h.Update)
	router.Delete("/:id", h.Delete)
}
