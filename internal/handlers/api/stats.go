package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// StatsHandler handles statistics requests
type StatsHandler struct{}

// NewStatsHandler creates new stats handler
func NewStatsHandler() *StatsHandler {
	return &StatsHandler{}
}

// Get returns dashboard statistics
// @Summary Get dashboard statistics
// @Tags stats
// @Produce json
// @Success 200 {object} map[string]any
// @Router /api/v1/stats [get]
func (h *StatsHandler) Get(c *fiber.Ctx) error {
	// Get user ID from auth context
	_, err := GetUserID(c)
	if err != nil {
		return err
	}

	// Get stats from database
	stats := map[string]any{
		"total_submissions":      0,
		"completed_submissions":  0,
		"pending_submissions":    0,
		"in_progress_submissions": 0,
	}

	// TODO: Implement actual stats queries
	// For now, return placeholder data

	return webutil.Response(c, fiber.StatusOK, "Stats retrieved", stats)
}

// RegisterRoutes registers stats routes
func (h *StatsHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/", h.Get)
}

