package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// EventHandler handles event requests
type EventHandler struct{}

// NewEventHandler creates new event handler
func NewEventHandler() *EventHandler {
	return &EventHandler{}
}

// List returns paginated list of events
// @Summary List events
// @Tags events
// @Param limit query int false "Limit" default(10)
// @Param sort query string false "Sort" default(created_at:desc)
// @Produce json
// @Success 200 {object} map[string]any
// @Router /api/v1/events [get]
func (h *EventHandler) List(c *fiber.Ctx) error {
	// Get pagination parameters
	limit := c.QueryInt("limit", 10)
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 10
	}

	// Get user ID from auth context
	_, err := GetUserID(c)
	if err != nil {
		return err
	}

	// TODO: Implement actual events query
	// For now, return empty array
	events := []models.Event{}

	return webutil.Response(c, fiber.StatusOK, "Events retrieved", map[string]any{
		"data":  events,
		"total": len(events),
		"limit": limit,
	})
}

// RegisterRoutes registers event routes
func (h *EventHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/", h.List)
}

