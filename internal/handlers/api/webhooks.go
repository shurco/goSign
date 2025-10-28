package api

import (
	"github.com/gofiber/fiber/v2"

	"github.com/shurco/gosign/internal/models"
)

// WebhookHandler handles requests to webhooks
type WebhookHandler struct {
	*ResourceHandler[models.Webhook] // embed generic CRUD
}

// NewWebhookHandler creates new handler
func NewWebhookHandler(repo ResourceRepository[models.Webhook]) *WebhookHandler {
	return &WebhookHandler{
		ResourceHandler: NewResourceHandler("webhook", repo),
	}
}

// RegisterRoutes registers all routes for webhooks
// Using only generic CRUD - no specific logic needed
func (h *WebhookHandler) RegisterRoutes(router fiber.Router) {
	h.ResourceHandler.RegisterRoutes(router)
}

