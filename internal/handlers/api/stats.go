package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// StatsHandler handles statistics requests
type StatsHandler struct {
	pool *pgxpool.Pool
}

// NewStatsHandler creates new stats handler
func NewStatsHandler(pool *pgxpool.Pool) *StatsHandler {
	return &StatsHandler{pool: pool}
}

// Get returns dashboard statistics
// @Summary Get dashboard statistics
// @Tags stats
// @Produce json
// @Success 200 {object} map[string]any
// @Router /api/v1/stats [get]
func (h *StatsHandler) Get(c *fiber.Ctx) error {
	// Get user ID from auth context
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	if h.pool == nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Stats service not initialized", nil)
	}

	var (
		totalSubmissions     int
		pendingSubmissions   int
		inProgressSubmissions int
		completedSubmissions int
		totalTemplates       int
		activeTemplates      int
		totalSubmitters      int
	)

	// KISS: derive submission status from submitter statuses.
	// This keeps logic consistent with /api/v1/signing-links list.
	err = h.pool.QueryRow(c.Context(), `
		WITH my_submissions AS (
			SELECT id
			FROM submission
			WHERE created_by_user_id = $1
		),
		per_submission AS (
			SELECT
				ms.id AS submission_id,
				CASE
					WHEN bool_and(COALESCE(s.status, 'pending') = 'completed') THEN 'completed'
					WHEN bool_or(COALESCE(s.status, 'pending') = 'declined') THEN 'declined'
					WHEN bool_or(COALESCE(s.status, 'pending') = 'opened')
						OR sum(CASE WHEN COALESCE(s.status, 'pending') = 'completed' THEN 1 ELSE 0 END) > 0
						THEN 'in_progress'
					ELSE 'pending'
				END AS status
			FROM my_submissions ms
			JOIN submitter s ON s.submission_id = ms.id
			GROUP BY ms.id
		)
		SELECT
			(SELECT count(*) FROM per_submission)::int AS total_submissions,
			(SELECT count(*) FROM per_submission WHERE status = 'pending')::int AS pending_submissions,
			(SELECT count(*) FROM per_submission WHERE status = 'in_progress')::int AS in_progress_submissions,
			(SELECT count(*) FROM per_submission WHERE status = 'completed')::int AS completed_submissions,
			(SELECT count(*) FROM template)::int AS total_templates,
			(SELECT count(*) FROM template WHERE archived_at IS NULL)::int AS active_templates,
			(SELECT count(*)
			 FROM submitter s
			 JOIN submission sub ON sub.id = s.submission_id
			 WHERE sub.created_by_user_id = $1
			)::int AS total_submitters
	`, userID).Scan(
		&totalSubmissions,
		&pendingSubmissions,
		&inProgressSubmissions,
		&completedSubmissions,
		&totalTemplates,
		&activeTemplates,
		&totalSubmitters,
	)
	if err != nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to load stats", nil)
	}

	stats := map[string]any{
		"total_submissions":       totalSubmissions,
		"pending_submissions":     pendingSubmissions,
		"in_progress_submissions": inProgressSubmissions,
		"completed_submissions":   completedSubmissions,
		"total_templates":         totalTemplates,
		"active_templates":        activeTemplates,
		"total_submitters":        totalSubmitters,
	}

	return webutil.Response(c, fiber.StatusOK, "Stats retrieved", stats)
}

// RegisterRoutes registers stats routes
func (h *StatsHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/", h.Get)
}

