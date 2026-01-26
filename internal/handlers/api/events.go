package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// EventHandler handles event requests
type EventHandler struct {
	pool *pgxpool.Pool
}

// NewEventHandler creates new event handler
func NewEventHandler(pool *pgxpool.Pool) *EventHandler {
	return &EventHandler{pool: pool}
}

type EventItem struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Message      string `json:"message"`
	DocumentName string `json:"document_name,omitempty"`
	CreatedAt    string `json:"created_at"`
	IP           string `json:"ip,omitempty"`
	Location     string `json:"location,omitempty"`
	Reason       string `json:"reason,omitempty"`
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
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	if h.pool == nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Events service not initialized", nil)
	}

	// If user is in an organization context, show org-wide activity by template ownership.
	// This matches how templates are scoped today (templates have organization_id; submissions don't).
	orgID, _ := GetOrganizationID(c)

	var rows pgx.Rows
	if orgID != "" {
		rows, err = h.pool.Query(c.Context(), `
			WITH scoped_submissions AS (
				SELECT
					sub.id,
					sub.created_at,
					COALESCE(t.name, '') AS document_name
				FROM submission sub
				JOIN template t ON t.id = sub.template_id
				WHERE t.organization_id = $1
			),
			activity AS (
				-- Submission created
				SELECT
					('submission_created:' || ss.id::text) AS id,
					'submission_created' AS type,
					ss.created_at AS ts,
					ss.document_name AS document_name,
					host(e_created.ip) AS ip,
					NULL AS location,
					NULL::text AS reason
				FROM scoped_submissions ss
				LEFT JOIN event e_created ON e_created.type = 'submission.created'
					AND e_created.resource_type = 'submission'
					AND e_created.resource_id = ss.id

				UNION ALL

				-- Submitter opened
				SELECT
					('submitter_opened:' || s.id::text) AS id,
					'submitter_opened' AS type,
					s.opened_at AS ts,
					ss.document_name AS document_name,
					host(e_opened.ip) AS ip,
					CASE 
						WHEN s.metadata->'location'->>'full' IS NOT NULL THEN s.metadata->'location'->>'full'
						WHEN s.metadata->>'location' IS NOT NULL AND jsonb_typeof(s.metadata->'location') = 'string' THEN s.metadata->>'location'
						ELSE NULL
					END AS location,
					NULL::text AS reason
				FROM submitter s
				JOIN scoped_submissions ss ON ss.id = s.submission_id
				LEFT JOIN event e_opened ON e_opened.type = 'submitter.opened'
					AND e_opened.resource_type = 'submission'
					AND e_opened.resource_id = ss.id
					AND e_opened.metadata_json->>'submitter_id' = s.id::text
				WHERE s.opened_at IS NOT NULL

				UNION ALL

				-- Submitter completed
				SELECT
					('submitter_completed:' || s.id::text) AS id,
					'submitter_completed' AS type,
					s.completed_at AS ts,
					ss.document_name AS document_name,
					host(e_completed.ip) AS ip,
					CASE 
						WHEN s.metadata->'location'->>'full' IS NOT NULL THEN s.metadata->'location'->>'full'
						WHEN s.metadata->>'location' IS NOT NULL AND jsonb_typeof(s.metadata->'location') = 'string' THEN s.metadata->>'location'
						ELSE NULL
					END AS location,
					NULL::text AS reason
				FROM submitter s
				JOIN scoped_submissions ss ON ss.id = s.submission_id
				LEFT JOIN event e_completed ON e_completed.type = 'submitter.completed'
					AND e_completed.resource_type = 'submission'
					AND e_completed.resource_id = ss.id
					AND e_completed.metadata_json->>'submitter_id' = s.id::text
				WHERE s.completed_at IS NOT NULL

				UNION ALL

				-- Submitter declined
				SELECT
					('submitter_declined:' || s.id::text) AS id,
					'submitter_declined' AS type,
					s.declined_at AS ts,
					ss.document_name AS document_name,
					host(e_declined.ip) AS ip,
					CASE 
						WHEN s.metadata->'location'->>'full' IS NOT NULL THEN s.metadata->'location'->>'full'
						WHEN s.metadata->>'location' IS NOT NULL AND jsonb_typeof(s.metadata->'location') = 'string' THEN s.metadata->>'location'
						ELSE NULL
					END AS location,
					s.metadata->>'decline_reason' AS reason
				FROM submitter s
				JOIN scoped_submissions ss ON ss.id = s.submission_id
				LEFT JOIN event e_declined ON e_declined.type = 'submitter.declined'
					AND e_declined.resource_type = 'submission'
					AND e_declined.resource_id = ss.id
					AND e_declined.metadata_json->>'submitter_id' = s.id::text
				WHERE s.declined_at IS NOT NULL
			)
			SELECT id, type, ts, document_name, ip, location, reason
			FROM activity
			ORDER BY ts DESC
			LIMIT $2
		`, orgID, limit)
} else {
		rows, err = h.pool.Query(c.Context(), `
			WITH scoped_submissions AS (
				SELECT
					sub.id,
					sub.created_at,
					COALESCE(t.name, '') AS document_name
				FROM submission sub
				JOIN template t ON t.id = sub.template_id
				WHERE sub.created_by_user_id = $1
			),
			activity AS (
				-- Submission created
				SELECT
					('submission_created:' || ss.id::text) AS id,
					'submission_created' AS type,
					ss.created_at AS ts,
					ss.document_name AS document_name,
					host(e_created.ip) AS ip,
					NULL AS location,
					NULL::text AS reason
				FROM scoped_submissions ss
				LEFT JOIN event e_created ON e_created.type = 'submission.created'
					AND e_created.resource_type = 'submission'
					AND e_created.resource_id = ss.id

				UNION ALL

				-- Submitter opened
				SELECT
					('submitter_opened:' || s.id::text) AS id,
					'submitter_opened' AS type,
					s.opened_at AS ts,
					ss.document_name AS document_name,
					host(e_opened.ip) AS ip,
					CASE 
						WHEN s.metadata->'location'->>'full' IS NOT NULL THEN s.metadata->'location'->>'full'
						WHEN s.metadata->>'location' IS NOT NULL AND jsonb_typeof(s.metadata->'location') = 'string' THEN s.metadata->>'location'
						ELSE NULL
					END AS location,
					NULL::text AS reason
				FROM submitter s
				JOIN scoped_submissions ss ON ss.id = s.submission_id
				LEFT JOIN event e_opened ON e_opened.type = 'submitter.opened'
					AND e_opened.resource_type = 'submission'
					AND e_opened.resource_id = ss.id
					AND e_opened.metadata_json->>'submitter_id' = s.id::text
				WHERE s.opened_at IS NOT NULL

				UNION ALL

				-- Submitter completed
				SELECT
					('submitter_completed:' || s.id::text) AS id,
					'submitter_completed' AS type,
					s.completed_at AS ts,
					ss.document_name AS document_name,
					host(e_completed.ip) AS ip,
					CASE 
						WHEN s.metadata->'location'->>'full' IS NOT NULL THEN s.metadata->'location'->>'full'
						WHEN s.metadata->>'location' IS NOT NULL AND jsonb_typeof(s.metadata->'location') = 'string' THEN s.metadata->>'location'
						ELSE NULL
					END AS location,
					NULL::text AS reason
				FROM submitter s
				JOIN scoped_submissions ss ON ss.id = s.submission_id
				LEFT JOIN event e_completed ON e_completed.type = 'submitter.completed'
					AND e_completed.resource_type = 'submission'
					AND e_completed.resource_id = ss.id
					AND e_completed.metadata_json->>'submitter_id' = s.id::text
				WHERE s.completed_at IS NOT NULL

				UNION ALL

				-- Submitter declined
				SELECT
					('submitter_declined:' || s.id::text) AS id,
					'submitter_declined' AS type,
					s.declined_at AS ts,
					ss.document_name AS document_name,
					host(e_declined.ip) AS ip,
					CASE 
						WHEN s.metadata->'location'->>'full' IS NOT NULL THEN s.metadata->'location'->>'full'
						WHEN s.metadata->>'location' IS NOT NULL AND jsonb_typeof(s.metadata->'location') = 'string' THEN s.metadata->>'location'
						ELSE NULL
					END AS location,
					s.metadata->>'decline_reason' AS reason
				FROM submitter s
				JOIN scoped_submissions ss ON ss.id = s.submission_id
				LEFT JOIN event e_declined ON e_declined.type = 'submitter.declined'
					AND e_declined.resource_type = 'submission'
					AND e_declined.resource_id = ss.id
					AND e_declined.metadata_json->>'submitter_id' = s.id::text
				WHERE s.declined_at IS NOT NULL
			)
			SELECT id, type, ts, document_name, ip, location, reason
			FROM activity
			ORDER BY ts DESC
			LIMIT $2
		`, userID, limit)
	}
	if err != nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to load events", nil)
	}
	defer rows.Close()

	items := make([]EventItem, 0)
	for rows.Next() {
		var (
			id        string
			typ       string
			createdAt time.Time
			docName   string
			ip        *string
			location  *string
			reason    *string
		)
		if err := rows.Scan(&id, &typ, &createdAt, &docName, &ip, &location, &reason); err != nil {
			return webutil.Response(c, fiber.StatusInternalServerError, "Failed to parse events", nil)
		}
		ipStr := ""
		if ip != nil {
			ipStr = *ip
		}
		locationStr := ""
		if location != nil {
			locationStr = *location
		}
		reasonStr := ""
		if reason != nil {
			reasonStr = *reason
		}
		items = append(items, EventItem{
			ID:           id,
			Type:         typ,
			Message:      eventMessage(typ),
			DocumentName: docName,
			CreatedAt:    createdAt.Format(time.RFC3339),
			IP:           ipStr,
			Location:     locationStr,
			Reason:       reasonStr,
		})
	}

	return webutil.Response(c, fiber.StatusOK, "Events retrieved", map[string]any{
		"items": items,
		"total": len(items),
		"limit": limit,
	})
}

func eventMessage(eventType string) string {
	switch eventType {
	case "submission_created":
		return "Signing created"
	case "submitter_opened":
		return "Signer opened the document"
	case "submitter_completed":
		return "Signer completed the document"
	case "submitter_declined":
		return "Signer declined the document"
	default:
		return eventType
	}
}

// RegisterRoutes registers event routes
func (h *EventHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/", h.List)
}

