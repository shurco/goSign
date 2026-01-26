package api

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/internal/services"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// SigningLinkHandler creates "direct link" signing flows (no email required).
// It is intentionally separate from generic CRUD because it needs to create
// both a submission and its submitter(s) and return the public signing URL.
type SigningLinkHandler struct {
	pool           *pgxpool.Pool
	templateQueries *queries.TemplateQueries
	completedDoc    *services.CompletedDocumentBuilder
}

func NewSigningLinkHandler(pool *pgxpool.Pool, templateQueries *queries.TemplateQueries, completedDoc *services.CompletedDocumentBuilder) *SigningLinkHandler {
	return &SigningLinkHandler{
		pool:            pool,
		templateQueries:  templateQueries,
		completedDoc:     completedDoc,
	}
}

type SubmitterInput struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}

type CreateSigningLinkRequest struct {
	TemplateID string `json:"template_id" validate:"required"`

	// Submitters count must match template submitters count.
	Submitters []SubmitterInput `json:"submitters" validate:"required,min=1"`

	// SigningMode: "sequential" or "parallel". Default "sequential".
	SigningMode string `json:"signing_mode,omitempty"`

	// Optional locale for the submission (used by i18n).
	Locale string `json:"locale,omitempty"`
}

type CreatedSubmitterLink struct {
	SubmitterID string `json:"submitter_id"`
	Slug        string `json:"slug"`
	DirectURL   string `json:"direct_url"` // "/s/:slug"
}

type CreateSigningLinkResponse struct {
	SubmissionID string                `json:"submission_id"`
	TemplateID   string                `json:"template_id"`
	Links        []CreatedSubmitterLink `json:"links"`
}

type ListSigningLinksItem struct {
	SubmissionID   string                 `json:"submission_id"`
	TemplateID     string                 `json:"template_id"`
	TemplateName   string                 `json:"template_name"`
	CreatedAt      string                 `json:"created_at"`
	Status         string                 `json:"status"`
	CompletedCount int                    `json:"completed_count"`
	TotalCount     int                    `json:"total_count"`
	Submitters     []map[string]any       `json:"submitters"`
	Links          []CreatedSubmitterLink `json:"links"`
}

type SigningLinkDetail struct {
	SubmissionID   string                 `json:"submission_id"`
	TemplateID     string                 `json:"template_id"`
	TemplateName   string                 `json:"template_name"`
	CreatedAt      string                 `json:"created_at"`
	CreatedIP      string                 `json:"created_ip,omitempty"`
	Status         string                 `json:"status"`
	CompletedCount int                    `json:"completed_count"`
	TotalCount     int                    `json:"total_count"`
	Submitters     []map[string]any       `json:"submitters"`
	Links          []CreatedSubmitterLink `json:"links"`
	DeclineEvents  []map[string]any       `json:"decline_events,omitempty"`
	OpenedEvents   []map[string]any       `json:"opened_events,omitempty"`
	CompletedEvents []map[string]any      `json:"completed_events,omitempty"`
}

// Create creates a new submission and N submitters (defined by template), and returns public signing URLs.
//
// @Summary Create direct signing link
// @Description Creates a submission without sending email and returns a unique signing link
// @Tags signing-links
// @Accept json
// @Produce json
// @Param body body CreateSigningLinkRequest true "Create signing link request"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/signing-links [post]
func (h *SigningLinkHandler) Create(c *fiber.Ctx) error {
	var req CreateSigningLinkRequest
	if err := parseAndValidateJSON(c, &req); err != nil {
		return err
	}

	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	ctx := c.Context()

	// Enforce signer count based on template definition.
	if h.templateQueries == nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Template queries not initialized", nil)
	}
	tpl, err := h.templateQueries.Template(ctx, req.TemplateID)
	if err != nil || tpl == nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Template not found", nil)
	}
	if len(tpl.Submitters) == 0 {
		return webutil.Response(c, fiber.StatusBadRequest, "Template has no submitters configured", nil)
	}
	if len(req.Submitters) != len(tpl.Submitters) {
		return webutil.Response(
			c,
			fiber.StatusBadRequest,
			fmt.Sprintf("Invalid submitters count: expected %d, got %d", len(tpl.Submitters), len(req.Submitters)),
			nil,
		)
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create signing link", nil)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	submissionID := uuid.NewString()
	submissionSlug := uuid.NewString()
	submittersOrder := "0"
	source := "direct_link"

	signingMode := req.SigningMode
	if signingMode != "sequential" && signingMode != "parallel" {
		signingMode = "sequential"
	}
	preferencesJSON := fmt.Sprintf(`{"signing_mode": %q}`, signingMode)

	// Note: DB schema uses created_by_user_id and requires slug/source/submitters_order.
	// Signing mode is stored in preferences for use by sequential/parallel flows.
	if req.Locale != "" {
		_, err = tx.Exec(ctx, `
			INSERT INTO submission (id, template_id, created_by_user_id, slug, source, submitters_order, locale, preferences)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb)
		`, submissionID, req.TemplateID, userID, submissionSlug, source, submittersOrder, req.Locale, preferencesJSON)
	} else {
		_, err = tx.Exec(ctx, `
			INSERT INTO submission (id, template_id, created_by_user_id, slug, source, submitters_order, preferences)
			VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
		`, submissionID, req.TemplateID, userID, submissionSlug, source, submittersOrder, preferencesJSON)
	}
	if err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, fmt.Sprintf("Failed to create submission: %v", err), nil)
	}

	links := make([]CreatedSubmitterLink, 0, len(req.Submitters))
	for i, s := range req.Submitters {
		submitterID := uuid.NewString()
		submitterSlug := uuid.NewString()

		meta := map[string]any{
			"template_submitter_id": tpl.Submitters[i].ID,
			"order":                i,
		}
		metaJSON, _ := json.Marshal(meta)

		_, err = tx.Exec(ctx, `
			INSERT INTO submitter (id, submission_id, name, email, phone, slug, metadata)
			VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), $6, $7::jsonb)
		`, submitterID, submissionID, s.Name, s.Email, s.Phone, submitterSlug, string(metaJSON))
		if err != nil {
			return webutil.Response(c, fiber.StatusBadRequest, fmt.Sprintf("Failed to create submitter: %v", err), nil)
		}

		links = append(links, CreatedSubmitterLink{
			SubmitterID: submitterID,
			Slug:        submitterSlug,
			DirectURL:   "/s/" + submitterSlug,
		})
	}

	// Record event for dashboard timeline (best-effort; in the same tx).
	clientIP := GetClientIP(c)
	_, _ = tx.Exec(ctx, `
		INSERT INTO event (id, type, actor_id, resource_type, resource_id, metadata_json, ip, created_at)
		VALUES (gen_random_uuid(), 'submission.created', $1, 'submission', $2, '{}'::jsonb, $3::inet, NOW())
	`, userID, submissionID, clientIP)

	if err := tx.Commit(ctx); err != nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create signing link", nil)
	}

	resp := CreateSigningLinkResponse{
		SubmissionID: submissionID,
		TemplateID:   req.TemplateID,
		Links:        links,
	}
	return webutil.Response(c, fiber.StatusCreated, "signing_link_created", resp)
}

// List returns submissions created via direct-link flow, including signer status and links.
// @Summary List direct-link signings
// @Tags signing-links
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} map[string]any
// @Router /api/v1/signing-links [get]
func (h *SigningLinkHandler) List(c *fiber.Ctx) error {
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 20)
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	rows, err := h.pool.Query(c.Context(), `
		SELECT
			sub.id AS submission_id,
			sub.template_id,
			COALESCE(t.name, '') AS template_name,
			sub.created_at::text AS created_at,
			CASE
				WHEN bool_and(COALESCE(s.status, 'pending') = 'completed') THEN 'completed'
				WHEN bool_or(COALESCE(s.status, 'pending') = 'declined') THEN 'declined'
				WHEN bool_or(COALESCE(s.status, 'pending') = 'opened')
					OR sum(CASE WHEN COALESCE(s.status, 'pending') = 'completed' THEN 1 ELSE 0 END) > 0
					THEN 'in_progress'
				ELSE 'pending'
			END AS status,
			sum(CASE WHEN COALESCE(s.status, 'pending') = 'completed' THEN 1 ELSE 0 END)::int AS completed_count,
			count(*)::int AS total_count,
			jsonb_agg(
				jsonb_build_object(
					'id', s.id,
					'name', COALESCE(s.name, ''),
					'email', COALESCE(s.email, ''),
					'phone', COALESCE(s.phone, ''),
					'slug', s.slug,
					'status', COALESCE(s.status, 'pending'),
					'completed_at', CASE WHEN s.completed_at IS NULL THEN NULL ELSE s.completed_at::text END
				)
				ORDER BY s.created_at ASC
			) AS submitters
		FROM submission sub
		JOIN template t ON t.id = sub.template_id
		JOIN submitter s ON s.submission_id = sub.id
		WHERE sub.created_by_user_id = $1
		  AND COALESCE(sub.source, '') = 'direct_link'
		GROUP BY sub.id, sub.template_id, t.name, sub.created_at
		ORDER BY sub.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, pageSize, offset)
	if err != nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to load signings", nil)
	}
	defer rows.Close()

	items := make([]ListSigningLinksItem, 0)
	for rows.Next() {
		var (
			submissionID   string
			templateID     string
			templateName   string
			createdAt      string
			status         string
			completedCount int
			totalCount     int
			submittersJSON []byte
		)
		if err := rows.Scan(&submissionID, &templateID, &templateName, &createdAt, &status, &completedCount, &totalCount, &submittersJSON); err != nil {
			return webutil.Response(c, fiber.StatusInternalServerError, "Failed to parse signings", nil)
		}

		var submitters []map[string]any
		_ = json.Unmarshal(submittersJSON, &submitters)

		links := make([]CreatedSubmitterLink, 0, len(submitters))
		for _, s := range submitters {
			slug, _ := s["slug"].(string)
			id, _ := s["id"].(string)
			if slug == "" {
				continue
			}
			links = append(links, CreatedSubmitterLink{
				SubmitterID: id,
				Slug:        slug,
				DirectURL:   "/s/" + slug,
			})
		}

		items = append(items, ListSigningLinksItem{
			SubmissionID:   submissionID,
			TemplateID:     templateID,
			TemplateName:   templateName,
			CreatedAt:      createdAt,
			Status:         status,
			CompletedCount: completedCount,
			TotalCount:     totalCount,
			Submitters:     submitters,
			Links:          links,
		})
	}

	return webutil.Response(c, fiber.StatusOK, "signing_links", map[string]any{
		"items":     items,
		"page":      page,
		"page_size": pageSize,
	})
}

// Get returns a single signing (submission) with signer details.
// It is used by the UI when opening a "status history" page.
//
// Access control:
// - If user is in organization context -> allow access to submissions created from org templates.
// - Otherwise -> allow access only to submissions created by the user.
//
// @Summary Get direct-link signing details
// @Tags signing-links
// @Produce json
// @Param submission_id path string true "Submission ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/signing-links/{submission_id} [get]
func (h *SigningLinkHandler) Get(c *fiber.Ctx) error {
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}
	orgID, _ := GetOrganizationID(c)

	submissionID := c.Params("submission_id")
	if submissionID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "submission_id is required", nil)
	}

	if h.pool == nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Signing links service not initialized", nil)
	}

	var (
		templateID     string
		templateName   string
		createdAt      string
		createdIP      *string
		status         string
		completedCount int
		totalCount     int
		submittersJSON []byte
	)

	if orgID != "" {
		err = h.pool.QueryRow(c.Context(), `
			SELECT
				sub.id AS submission_id,
				sub.template_id,
				COALESCE(t.name, '') AS template_name,
				sub.created_at::text AS created_at,
				host(created_event.ip) AS created_ip,
				CASE
					WHEN bool_and(COALESCE(s.status, 'pending') = 'completed') THEN 'completed'
					WHEN bool_or(COALESCE(s.status, 'pending') = 'declined') THEN 'declined'
					WHEN bool_or(COALESCE(s.status, 'pending') = 'opened')
						OR sum(CASE WHEN COALESCE(s.status, 'pending') = 'completed' THEN 1 ELSE 0 END) > 0
						THEN 'in_progress'
					ELSE 'pending'
				END AS status,
				sum(CASE WHEN COALESCE(s.status, 'pending') = 'completed' THEN 1 ELSE 0 END)::int AS completed_count,
				count(*)::int AS total_count,
				jsonb_agg(
					jsonb_build_object(
						'id', s.id,
						'name', COALESCE(s.name, ''),
						'email', COALESCE(s.email, ''),
						'phone', COALESCE(s.phone, ''),
						'slug', s.slug,
						'status', COALESCE(s.status, 'pending'),
						'created_at', s.created_at::text,
						'opened_at', CASE WHEN s.opened_at IS NULL THEN NULL ELSE s.opened_at::text END,
						'opened_ip', host(opened_event.ip),
						'opened_location', CASE 
							WHEN s.metadata->'location'->>'full' IS NOT NULL THEN s.metadata->'location'->>'full'
							WHEN s.metadata->>'location' IS NOT NULL AND jsonb_typeof(s.metadata->'location') = 'string' THEN s.metadata->>'location'
							ELSE NULL
						END,
						'completed_at', CASE WHEN s.completed_at IS NULL THEN NULL ELSE s.completed_at::text END,
						'completed_ip', host(completed_event.ip),
						'completed_location', CASE 
							WHEN s.metadata->'location'->>'full' IS NOT NULL THEN s.metadata->'location'->>'full'
							WHEN s.metadata->>'location' IS NOT NULL AND jsonb_typeof(s.metadata->'location') = 'string' THEN s.metadata->>'location'
							ELSE NULL
						END,
						'declined_at', CASE WHEN s.declined_at IS NULL THEN NULL ELSE s.declined_at::text END,
						'declined_ip', host(declined_event.ip),
						'declined_location', CASE 
							WHEN s.metadata->'location'->>'full' IS NOT NULL THEN s.metadata->'location'->>'full'
							WHEN s.metadata->>'location' IS NOT NULL AND jsonb_typeof(s.metadata->'location') = 'string' THEN s.metadata->>'location'
							ELSE NULL
						END,
						'decline_reason', s.metadata->>'decline_reason'
					)
					ORDER BY s.created_at ASC
				) AS submitters
			FROM submission sub
			JOIN template t ON t.id = sub.template_id
			JOIN submitter s ON s.submission_id = sub.id
			LEFT JOIN event created_event ON created_event.type = 'submission.created'
				AND created_event.resource_type = 'submission'
				AND created_event.resource_id = sub.id
			LEFT JOIN LATERAL (
				SELECT e.ip FROM event e
				WHERE e.type = 'submitter.opened' AND e.resource_type = 'submission' AND e.resource_id = sub.id AND e.metadata_json->>'submitter_id' = s.id::text
				ORDER BY e.created_at DESC LIMIT 1
			) opened_event ON true
			LEFT JOIN LATERAL (
				SELECT e.ip FROM event e
				WHERE e.type = 'submitter.completed' AND e.resource_type = 'submission' AND e.resource_id = sub.id AND e.metadata_json->>'submitter_id' = s.id::text
				ORDER BY e.created_at DESC LIMIT 1
			) completed_event ON true
			LEFT JOIN LATERAL (
				SELECT e.ip FROM event e
				WHERE e.type = 'submitter.declined' AND e.resource_type = 'submission' AND e.resource_id = sub.id AND e.metadata_json->>'submitter_id' = s.id::text
				ORDER BY e.created_at DESC LIMIT 1
			) declined_event ON true
			WHERE sub.id = $1
			  AND COALESCE(sub.source, '') = 'direct_link'
			  AND t.organization_id = $2
			GROUP BY sub.id, sub.template_id, t.name, sub.created_at, created_event.ip
			LIMIT 1
		`, submissionID, orgID).Scan(&submissionID, &templateID, &templateName, &createdAt, &createdIP, &status, &completedCount, &totalCount, &submittersJSON)
	} else {
		err = h.pool.QueryRow(c.Context(), `
			SELECT
				sub.id AS submission_id,
				sub.template_id,
				COALESCE(t.name, '') AS template_name,
				sub.created_at::text AS created_at,
				host(created_event.ip) AS created_ip,
				CASE
					WHEN bool_and(COALESCE(s.status, 'pending') = 'completed') THEN 'completed'
					WHEN bool_or(COALESCE(s.status, 'pending') = 'declined') THEN 'declined'
					WHEN bool_or(COALESCE(s.status, 'pending') = 'opened')
						OR sum(CASE WHEN COALESCE(s.status, 'pending') = 'completed' THEN 1 ELSE 0 END) > 0
						THEN 'in_progress'
					ELSE 'pending'
				END AS status,
				sum(CASE WHEN COALESCE(s.status, 'pending') = 'completed' THEN 1 ELSE 0 END)::int AS completed_count,
				count(*)::int AS total_count,
				jsonb_agg(
					jsonb_build_object(
						'id', s.id,
						'name', COALESCE(s.name, ''),
						'email', COALESCE(s.email, ''),
						'phone', COALESCE(s.phone, ''),
						'slug', s.slug,
						'status', COALESCE(s.status, 'pending'),
						'created_at', s.created_at::text,
						'opened_at', CASE WHEN s.opened_at IS NULL THEN NULL ELSE s.opened_at::text END,
						'opened_ip', host(opened_event.ip),
						'opened_location', CASE 
							WHEN s.metadata->'location'->>'full' IS NOT NULL THEN s.metadata->'location'->>'full'
							WHEN s.metadata->>'location' IS NOT NULL AND jsonb_typeof(s.metadata->'location') = 'string' THEN s.metadata->>'location'
							ELSE NULL
						END,
						'completed_at', CASE WHEN s.completed_at IS NULL THEN NULL ELSE s.completed_at::text END,
						'completed_ip', host(completed_event.ip),
						'completed_location', CASE 
							WHEN s.metadata->'location'->>'full' IS NOT NULL THEN s.metadata->'location'->>'full'
							WHEN s.metadata->>'location' IS NOT NULL AND jsonb_typeof(s.metadata->'location') = 'string' THEN s.metadata->>'location'
							ELSE NULL
						END,
						'declined_at', CASE WHEN s.declined_at IS NULL THEN NULL ELSE s.declined_at::text END,
						'declined_ip', host(declined_event.ip),
						'declined_location', CASE 
							WHEN s.metadata->'location'->>'full' IS NOT NULL THEN s.metadata->'location'->>'full'
							WHEN s.metadata->>'location' IS NOT NULL AND jsonb_typeof(s.metadata->'location') = 'string' THEN s.metadata->>'location'
							ELSE NULL
						END,
						'decline_reason', s.metadata->>'decline_reason'
					)
					ORDER BY s.created_at ASC
				) AS submitters
			FROM submission sub
			JOIN template t ON t.id = sub.template_id
			JOIN submitter s ON s.submission_id = sub.id
			LEFT JOIN event created_event ON created_event.type = 'submission.created'
				AND created_event.resource_type = 'submission'
				AND created_event.resource_id = sub.id
			LEFT JOIN LATERAL (
				SELECT e.ip FROM event e
				WHERE e.type = 'submitter.opened' AND e.resource_type = 'submission' AND e.resource_id = sub.id AND e.metadata_json->>'submitter_id' = s.id::text
				ORDER BY e.created_at DESC LIMIT 1
			) opened_event ON true
			LEFT JOIN LATERAL (
				SELECT e.ip FROM event e
				WHERE e.type = 'submitter.completed' AND e.resource_type = 'submission' AND e.resource_id = sub.id AND e.metadata_json->>'submitter_id' = s.id::text
				ORDER BY e.created_at DESC LIMIT 1
			) completed_event ON true
			LEFT JOIN LATERAL (
				SELECT e.ip FROM event e
				WHERE e.type = 'submitter.declined' AND e.resource_type = 'submission' AND e.resource_id = sub.id AND e.metadata_json->>'submitter_id' = s.id::text
				ORDER BY e.created_at DESC LIMIT 1
			) declined_event ON true
			WHERE sub.id = $1
			  AND COALESCE(sub.source, '') = 'direct_link'
			  AND sub.created_by_user_id = $2
			GROUP BY sub.id, sub.template_id, t.name, sub.created_at, created_event.ip
			LIMIT 1
		`, submissionID, userID).Scan(&submissionID, &templateID, &templateName, &createdAt, &createdIP, &status, &completedCount, &totalCount, &submittersJSON)
	}
	if err != nil {
		return webutil.Response(c, fiber.StatusNotFound, "Signing not found", nil)
	}

	var submitters []map[string]any
	_ = json.Unmarshal(submittersJSON, &submitters)

	links := make([]CreatedSubmitterLink, 0, len(submitters))
	for _, s := range submitters {
		slug, _ := s["slug"].(string)
		id, _ := s["id"].(string)
		if slug == "" {
			continue
		}
		links = append(links, CreatedSubmitterLink{
			SubmitterID: id,
			Slug:        slug,
			DirectURL:   "/s/" + slug,
		})
	}

	openedEvents := make([]map[string]any, 0)
	if rowsOpened, errOpen := h.pool.Query(c.Context(), `
		SELECT e.created_at::text, e.metadata_json->>'submitter_id', COALESCE(s.name, ''), host(e.ip)
		FROM event e
		LEFT JOIN submitter s ON s.id = (e.metadata_json->>'submitter_id')::uuid AND s.submission_id = $1::uuid
		WHERE e.type = 'submitter.opened' AND e.resource_type = 'submission' AND e.resource_id = $1::uuid
		ORDER BY e.created_at ASC
	`, submissionID, submissionID); errOpen == nil {
		for rowsOpened.Next() {
			var at, subID, name string
			var ip *string
			if rowsOpened.Scan(&at, &subID, &name, &ip) == nil {
				ipStr := ""
				if ip != nil {
					ipStr = *ip
				}
				openedEvents = append(openedEvents, map[string]any{
					"at":             at,
					"submitter_id":   subID,
					"submitter_name": name,
					"ip":             ipStr,
				})
			}
		}
		rowsOpened.Close()
	}

	completedEvents := make([]map[string]any, 0)
	if rowsCompleted, errComp := h.pool.Query(c.Context(), `
		SELECT e.created_at::text, e.metadata_json->>'submitter_id', COALESCE(s.name, ''), host(e.ip)
		FROM event e
		LEFT JOIN submitter s ON s.id = (e.metadata_json->>'submitter_id')::uuid AND s.submission_id = $1::uuid
		WHERE e.type = 'submitter.completed' AND e.resource_type = 'submission' AND e.resource_id = $1::uuid
		ORDER BY e.created_at ASC
	`, submissionID, submissionID); errComp == nil {
		for rowsCompleted.Next() {
			var at, subID, name string
			var ip *string
			if rowsCompleted.Scan(&at, &subID, &name, &ip) == nil {
				ipStr := ""
				if ip != nil {
					ipStr = *ip
				}
				completedEvents = append(completedEvents, map[string]any{
					"at":             at,
					"submitter_id":   subID,
					"submitter_name": name,
					"ip":             ipStr,
				})
			}
		}
		rowsCompleted.Close()
	}

	declineEvents := make([]map[string]any, 0)
	rowsDecline, errDecline := h.pool.Query(c.Context(), `
		SELECT e.created_at::text, e.metadata_json->>'submitter_id', COALESCE(s.name, ''), host(e.ip), e.metadata_json->>'reason'
		FROM event e
		LEFT JOIN submitter s ON s.id = (e.metadata_json->>'submitter_id')::uuid AND s.submission_id = $1::uuid
		WHERE e.type = 'submitter.declined' AND e.resource_type = 'submission' AND e.resource_id = $1::uuid
		ORDER BY e.created_at ASC
	`, submissionID, submissionID)
	if errDecline == nil {
		defer rowsDecline.Close()
		for rowsDecline.Next() {
			var at, subID, name string
			var ip, reason *string
			if rowsDecline.Scan(&at, &subID, &name, &ip, &reason) == nil {
				ipStr := ""
				if ip != nil {
					ipStr = *ip
				}
				reasonStr := ""
				if reason != nil {
					reasonStr = *reason
				}
				declineEvents = append(declineEvents, map[string]any{
					"at":             at,
					"submitter_id":   subID,
					"submitter_name": name,
					"ip":             ipStr,
					"reason":         reasonStr,
				})
			}
		}
	}

	createdIPStr := ""
	if createdIP != nil {
		createdIPStr = *createdIP
	}
	detail := SigningLinkDetail{
		SubmissionID:   submissionID,
		TemplateID:     templateID,
		TemplateName:   templateName,
		CreatedAt:      createdAt,
		CreatedIP:      createdIPStr,
		Status:         status,
		CompletedCount: completedCount,
		TotalCount:     totalCount,
		Submitters:     submitters,
		Links:            links,
		DeclineEvents:    declineEvents,
		OpenedEvents:     openedEvents,
		CompletedEvents:  completedEvents,
	}
	return webutil.Response(c, fiber.StatusOK, "signing_link", detail)
}

// DownloadCompletedDocument downloads the final PDF for a completed submission.
// Only the creator of the submission can download it.
//
// @Summary Download completed document
// @Tags signing-links
// @Produce application/pdf
// @Param submission_id path string true "Submission ID"
// @Success 200 {file} file
// @Failure 403 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Router /api/v1/signing-links/{submission_id}/document [get]
func (h *SigningLinkHandler) DownloadCompletedDocument(c *fiber.Ctx) error {
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	submissionID := c.Params("submission_id")
	if submissionID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "submission_id is required", nil)
	}
	if h.completedDoc == nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Document builder not configured", nil)
	}

	// Ensure ownership.
	var ok bool
	err = h.pool.QueryRow(c.Context(), `
		SELECT EXISTS(
			SELECT 1
			FROM submission
			WHERE id = $1
			  AND created_by_user_id = $2
		)
	`, submissionID, userID).Scan(&ok)
	if err != nil || !ok {
		return webutil.Response(c, fiber.StatusNotFound, "Submission not found", nil)
	}

	isDone, err := h.completedDoc.IsSubmissionFullyCompleted(c.Context(), submissionID)
	if err != nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to check completion", nil)
	}
	if !isDone {
		return webutil.Response(c, fiber.StatusConflict, "Submission not completed yet", nil)
	}

	// Ensure we have an absolute base URL stored for QR codes in the certificate.
	baseURL := fmt.Sprintf("%s://%s", c.Protocol(), c.Get("Host"))
	_, _ = h.pool.Exec(c.Context(), `
		UPDATE submission
		SET preferences = jsonb_set(COALESCE(preferences, '{}'::jsonb), '{public_base_url}', to_jsonb($2::text), true),
		    updated_at = NOW()
		WHERE id = $1
		  AND COALESCE(preferences->>'public_base_url', '') = ''
	`, submissionID, baseURL)

	path, err := h.completedDoc.EnsureCompletedPDF(c.Context(), submissionID)
	if err != nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to build completed document", map[string]any{"error": err.Error()})
	}

	return c.Download(path, fmt.Sprintf("submission_%s.pdf", submissionID))
}

// parseAndValidateJSON is a small helper to keep handler code minimal.
func parseAndValidateJSON(c *fiber.Ctx, v interface{}) error {
	if err := c.BodyParser(v); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if err := webutil.ValidateStruct(v); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	return nil
}

