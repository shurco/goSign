package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/internal/services"
	"github.com/shurco/gosign/pkg/geolocation"
	"github.com/shurco/gosign/pkg/notification"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// PublicSigningHandler exposes public (no-auth) endpoints for signing by slug.
// This is used by the signer-facing UI at /s/:slug.
type PublicSigningHandler struct {
	pool            *pgxpool.Pool
	templateQueries  *queries.TemplateQueries
	userQueries     *queries.UserQueries
	notificationSvc  *notification.Service
	completedDoc     *services.CompletedDocumentBuilder
	geolocationSvc   *geolocation.Service
}

func NewPublicSigningHandler(
	pool *pgxpool.Pool,
	templateQueries *queries.TemplateQueries,
	userQueries *queries.UserQueries,
	notificationSvc *notification.Service,
	completedDoc *services.CompletedDocumentBuilder,
	geolocationSvc *geolocation.Service,
) *PublicSigningHandler {
	return &PublicSigningHandler{
		pool:            pool,
		templateQueries:  templateQueries,
		userQueries:     userQueries,
		notificationSvc:  notificationSvc,
		completedDoc:     completedDoc,
		geolocationSvc:   geolocationSvc,
	}
}

type getBySlugResponse struct {
	Template            *models.Template  `json:"template"`
	Submitter           *models.Submitter `json:"submitter"`
	SubmissionStatus    string            `json:"submission_status"`
	CompletedDocumentURL string           `json:"completed_document_url,omitempty"`
}

// GetBySlug returns template + submitter data for the signing portal.
// @Summary Get signing data by slug
// @Description Returns submitter and resolved template data for public signing UI
// @Tags public-signing
// @Produce json
// @Param slug path string true "Submitter slug"
// @Success 200 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Router /public/sign/{slug} [get]
func (h *PublicSigningHandler) GetBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return webutil.Response(c, fiber.StatusNotFound, "Not found", nil)
	}

	ctx := c.Context()

	// Fetch submitter + template_id (via submission).
	var (
		submitterID   string
		submissionID  string
		name          string
		email         string
		phone         string
		status        string
		completedAt   *time.Time
		declinedAt    *time.Time
		openedAt      *time.Time
		updatedAt     time.Time
		templateID    string
		metaJSONString string
	)

	err := h.pool.QueryRow(ctx, `
		SELECT
			s.id,
			s.submission_id,
			COALESCE(s.name, ''),
			COALESCE(s.email, ''),
			COALESCE(s.phone, ''),
			COALESCE(s.status, 'pending') AS status,
			s.completed_at,
			s.declined_at,
			s.opened_at,
			s.updated_at,
			sub.template_id,
			COALESCE(s.metadata, '{}'::jsonb)::text AS metadata_json
		FROM submitter s
		JOIN submission sub ON sub.id = s.submission_id
		WHERE s.slug = $1
		LIMIT 1
	`, slug).Scan(
		&submitterID,
		&submissionID,
		&name,
		&email,
		&phone,
		&status,
		&completedAt,
		&declinedAt,
		&openedAt,
		&updatedAt,
		&templateID,
		&metaJSONString,
	)
	if err != nil {
		return webutil.Response(c, fiber.StatusNotFound, "Submission not found", nil)
	}

	tpl, err := h.templateQueries.Template(ctx, templateID)
	if err != nil || tpl == nil {
		return webutil.Response(c, fiber.StatusNotFound, "Template not found", nil)
	}

	// Compute overall submission status (based on all submitters).
	submissionStatus := "pending"
	_ = h.pool.QueryRow(ctx, `
		SELECT CASE
			WHEN bool_and(COALESCE(s.status, 'pending') = 'completed') THEN 'completed'
			WHEN bool_or(COALESCE(s.status, 'pending') = 'declined') THEN 'declined'
			WHEN bool_or(COALESCE(s.status, 'pending') = 'opened')
				OR sum(CASE WHEN COALESCE(s.status, 'pending') = 'completed' THEN 1 ELSE 0 END) > 0
				THEN 'in_progress'
			ELSE 'pending'
		END AS status
		FROM submitter s
		WHERE s.submission_id = $1
		GROUP BY s.submission_id
		LIMIT 1
	`, submissionID).Scan(&submissionStatus)

	// Build submitter model. For public UI we only need a subset, but keep it consistent.
	submitter := &models.Submitter{
		ID:           submitterID,
		Name:         name,
		Email:        email,
		Phone:        phone,
		Slug:         slug,
		Status:       models.SubmitterStatus(status),
		SubmissionID: submissionID,
		CompletedAt:  completedAt,
		DeclinedAt:   declinedAt,
		OpenedAt:     openedAt,
	}

	// Backfill timestamps for legacy rows where status was set without *_at.
	// This prevents the signer UI from showing empty "Completed on:" / "Declined on:".
	switch submitter.Status {
	case models.SubmitterStatusCompleted:
		if submitter.CompletedAt == nil {
			submitter.CompletedAt = &updatedAt
		}
	case models.SubmitterStatusDeclined:
		if submitter.DeclinedAt == nil {
			submitter.DeclinedAt = &updatedAt
		}
	case models.SubmitterStatusOpened:
		if submitter.OpenedAt == nil {
			submitter.OpenedAt = &updatedAt
		}
	}

	// Resolve template fields for this particular submitter.
	// - Multi-signer templates have fields assigned to template submitter IDs.
	// - We store the mapping in submitter.metadata.template_submitter_id when creating the signing.
	// - For legacy/single-signer flows (no mapping), we fallback to "assign all fields to this signer".
	var meta map[string]any
	_ = json.Unmarshal([]byte(metaJSONString), &meta)
	templateSubmitterID, _ := meta["template_submitter_id"].(string)

	if templateSubmitterID != "" {
		for i := range tpl.Fields {
			if tpl.Fields[i].SubmitterID == templateSubmitterID {
				tpl.Fields[i].SubmitterID = submitter.ID
			}
		}
	} else {
		for i := range tpl.Fields {
			tpl.Fields[i].SubmitterID = submitter.ID
		}
	}

	resp := getBySlugResponse{
		Template:         tpl,
		Submitter:        submitter,
		SubmissionStatus: submissionStatus,
	}
	if submissionStatus == "completed" {
		resp.CompletedDocumentURL = fmt.Sprintf("/public/sign/%s/document", slug)
	}

	return webutil.Response(c, fiber.StatusOK, "ok", resp)
}

type completeRequest struct {
	Fields map[string]any `json:"fields" validate:"required"`
}

// Open marks the submitter as opened (best-effort).
// @Summary Mark submitter opened
// @Tags public-signing
// @Param slug path string true "Submitter slug"
// @Success 200 {object} map[string]any
// @Router /public/sign/{slug}/open [post]
func (h *PublicSigningHandler) Open(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return webutil.Response(c, fiber.StatusNotFound, "Not found", nil)
	}

	// Also record an event for the submission dashboard (best-effort).
	clientIP := getClientIP(c)
	_, _ = h.pool.Exec(c.Context(), `
		WITH upd AS (
			UPDATE submitter
			SET opened_at = COALESCE(opened_at, NOW()),
			    status = CASE WHEN status = 'pending' THEN 'opened' ELSE status END,
			    updated_at = NOW(),
			    ip = COALESCE(ip, $2::inet)
			WHERE slug = $1
			RETURNING id, submission_id
		)
		INSERT INTO event (id, type, resource_type, resource_id, metadata_json, ip, created_at)
		SELECT gen_random_uuid(), 'submitter.opened', 'submission', submission_id,
		       jsonb_build_object('submitter_id', id), $2::inet, NOW()
		FROM upd
	`, slug, clientIP)

	return webutil.Response(c, fiber.StatusOK, "opened", map[string]any{"slug": slug})
}

type updateSubmitterRequest struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required"`
}

// UpdateSubmitter updates submitter with email and name, links to user account if exists, and sends confirmation.
// @Summary Update submitter info
// @Tags public-signing
// @Accept json
// @Produce json
// @Param slug path string true "Submitter slug"
// @Param body body updateSubmitterRequest true "Update payload"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Router /public/sign/{slug}/update [post]
func (h *PublicSigningHandler) UpdateSubmitter(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return webutil.Response(c, fiber.StatusNotFound, "Not found", nil)
	}

	var req updateSubmitterRequest
	if err := parseAndValidate(c, &req); err != nil {
		// parseAndValidate already returns JSON response
		return err
	}

	// Validate email format on backend as well
	if req.Email == "" || req.Name == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Email and name are required", nil)
	}

	ctx := c.Context()
	baseURL := fmt.Sprintf("%s://%s", c.Protocol(), c.Get("Host"))

	// Update submitter and get submission_id
	var submissionID string
	var submitterID string
	err := h.pool.QueryRow(ctx, `
		UPDATE submitter
		SET name = $2,
		    email = $3,
		    updated_at = NOW()
		WHERE slug = $1
		  AND (COALESCE(name, '') = '' OR COALESCE(email, '') = '')
		RETURNING id, submission_id
	`, slug, req.Name, req.Email).Scan(&submitterID, &submissionID)
	if err != nil {
		log.Warn().Err(err).Str("slug", slug).Msg("Failed to update submitter")
		return webutil.Response(c, fiber.StatusNotFound, "Submitter not found or already has email/name", nil)
	}
	if submissionID == "" || submitterID == "" {
		return webutil.Response(c, fiber.StatusNotFound, "Submitter not found or already has email/name", nil)
	}

	// Check if user exists with this email
	if h.userQueries != nil {
		user, err := h.userQueries.GetUserByEmail(ctx, req.Email)
		if err == nil && user != nil {
			// Link submission to user account
			_, execErr := h.pool.Exec(ctx, `
				UPDATE submission
				SET created_by_user_id = $2,
				    account_id = $3,
				    updated_at = NOW()
				WHERE id = $1
				  AND created_by_user_id IS NULL
			`, submissionID, user.ID, user.AccountID)
			if execErr != nil {
				log.Warn().Err(execErr).Str("submission_id", submissionID).Str("user_id", user.ID).Msg("Failed to link submission to user account")
			}
		}
		// Ignore error if user not found - that's expected
	}

	// Send confirmation email (best-effort)
	if h.notificationSvc != nil && h.notificationSvc.CanSend(models.NotificationTypeEmail) {
		signingURL := fmt.Sprintf("%s/s/%s", strings.TrimRight(baseURL, "/"), slug)
		n := &models.Notification{
			ID:        uuid.NewString(),
			Type:      models.NotificationTypeEmail,
			Recipient: req.Email,
			Subject:   "Document signing invitation",
			Body:      fmt.Sprintf("Hello %s,\n\nYou have been invited to sign a document.\n\nSign here: %s\n", req.Name, signingURL),
			Context:   map[string]any{},
		}
		if err := h.notificationSvc.Send(n); err != nil {
			log.Warn().Err(err).Str("email", req.Email).Msg("Failed to send confirmation email")
		}
	}

	return webutil.Response(c, fiber.StatusOK, "updated", map[string]any{"slug": slug})
}

// Complete stores field values and marks submitter completed.
// @Summary Complete signing
// @Tags public-signing
// @Accept json
// @Produce json
// @Param slug path string true "Submitter slug"
// @Param body body completeRequest true "Fields payload"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Router /public/sign/{slug}/complete [post]
func (h *PublicSigningHandler) Complete(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return webutil.Response(c, fiber.StatusNotFound, "Not found", nil)
	}

	var req completeRequest
	if err := parseAndValidate(c, &req); err != nil {
		return err
	}

	// Mark as completed and return ids (single statement).
	clientIP := getClientIP(c)
	
	// Determine location from IP address (saved once, used later for certificate)
	var locationData map[string]any
	if h.geolocationSvc != nil && clientIP != "" {
		loc := h.geolocationSvc.GetLocation(clientIP)
		if loc != nil && loc.Full != "" {
			locationData = map[string]any{
				"city":    loc.City,
				"country": loc.Country,
				"full":    loc.Full,
			}
			log.Debug().Str("ip", clientIP).Str("location", loc.Full).Msg("Location determined for IP")
		} else {
			log.Debug().Str("ip", clientIP).Msg("Location not found for IP")
		}
	} else {
		if h.geolocationSvc == nil {
			log.Debug().Msg("Geolocation service not initialized")
		}
		if clientIP == "" {
			log.Debug().Msg("Client IP is empty")
		}
	}
	
	// Build metadata with fields and location
	metadataUpdates := map[string]any{
		"fields": req.Fields,
	}
	if locationData != nil {
		metadataUpdates["location"] = locationData
	}
	metadataJSON, err := json.Marshal(metadataUpdates)
	if err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid fields payload", nil)
	}
	
	var submissionID string
	var submitterID string
	err = h.pool.QueryRow(c.Context(), `
		WITH upd AS (
			UPDATE submitter
			SET metadata = COALESCE(metadata, '{}'::jsonb) || $2::jsonb,
			    status = 'completed',
			    completed_at = NOW(),
			    updated_at = NOW(),
			    ip = $3::inet
			WHERE slug = $1
			  AND COALESCE(status, 'pending') <> 'completed'
			RETURNING id, submission_id
		), ev AS (
			INSERT INTO event (id, type, resource_type, resource_id, metadata_json, ip, created_at)
			SELECT gen_random_uuid(), 'submitter.completed', 'submission', submission_id,
			       jsonb_build_object('submitter_id', id), $3::inet, NOW()
			FROM upd
		)
		SELECT submission_id, id
		FROM upd
		LIMIT 1
	`, slug, string(metadataJSON), clientIP).Scan(&submissionID, &submitterID)
	if err != nil || submissionID == "" || submitterID == "" {
		return webutil.Response(c, fiber.StatusNotFound, "Submitter not found", nil)
	}

	// Best-effort finalization (generate completed PDF + auto-send links).
	// Uses a DB idempotency flag in submission.preferences, so concurrent completions won't double-send.
	baseURL := fmt.Sprintf("%s://%s", c.Protocol(), c.Get("Host"))
	ctxAsync, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	go func() {
		defer cancel()
		h.finalizeIfCompleted(ctxAsync, submissionID, baseURL)
	}()

	return webutil.Response(c, fiber.StatusOK, "completed", map[string]any{"slug": slug})
}

type declineRequest struct {
	Reason string `json:"reason,omitempty"`
}

// Decline marks submitter declined.
// @Summary Decline signing
// @Tags public-signing
// @Accept json
// @Produce json
// @Param slug path string true "Submitter slug"
// @Param body body declineRequest false "Decline payload"
// @Success 200 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Router /public/sign/{slug}/decline [post]
func (h *PublicSigningHandler) Decline(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return webutil.Response(c, fiber.StatusNotFound, "Not found", nil)
	}

	var req declineRequest
	_ = c.BodyParser(&req) // optional

	clientIP := getClientIP(c)
	tag, err := h.pool.Exec(c.Context(), `
		WITH upd AS (
			UPDATE submitter
			SET metadata = COALESCE(metadata, '{}'::jsonb) || jsonb_build_object('decline_reason', NULLIF($2, '')),
			    status = 'declined',
			    declined_at = NOW(),
			    updated_at = NOW()
			WHERE slug = $1
			  AND status <> 'declined'
			RETURNING id, submission_id
		)
		INSERT INTO event (id, type, resource_type, resource_id, metadata_json, ip, created_at)
		SELECT gen_random_uuid(), 'submitter.declined', 'submission', submission_id,
		       jsonb_build_object('submitter_id', id, 'reason', NULLIF($2, '')), $3::inet, NOW()
		FROM upd
	`, slug, req.Reason, clientIP)
	if err != nil {
		return webutil.Response(c, fiber.StatusInternalServerError, fmt.Sprintf("Failed to decline: %v", err), nil)
	}
	if tag.RowsAffected() == 0 {
		return webutil.Response(c, fiber.StatusNotFound, "Submitter not found", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "declined", map[string]any{"slug": slug})
}

// GetCompletedDocument returns the final PDF only when the whole submission is completed.
// Public access is protected by submitter slug entropy.
func (h *PublicSigningHandler) GetCompletedDocument(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return webutil.Response(c, fiber.StatusNotFound, "Not found", nil)
	}
	if h.completedDoc == nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Document builder not configured", nil)
	}

	var submissionID string
	err := h.pool.QueryRow(c.Context(), `
		SELECT submission_id
		FROM submitter
		WHERE slug = $1
		LIMIT 1
	`, slug).Scan(&submissionID)
	if err != nil || submissionID == "" {
		return webutil.Response(c, fiber.StatusNotFound, "Submitter not found", nil)
	}

	ok, err := h.completedDoc.IsSubmissionFullyCompleted(c.Context(), submissionID)
	if err != nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to check completion", nil)
	}
	if !ok {
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

// GetCertificate returns the certificate PDF only when the whole submission is completed.
// Public access is protected by submitter slug entropy.
func (h *PublicSigningHandler) GetCertificate(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return webutil.Response(c, fiber.StatusNotFound, "Not found", nil)
	}
	if h.completedDoc == nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Document builder not configured", nil)
	}

	var submissionID string
	err := h.pool.QueryRow(c.Context(), `
		SELECT submission_id
		FROM submitter
		WHERE slug = $1
		LIMIT 1
	`, slug).Scan(&submissionID)
	if err != nil || submissionID == "" {
		return webutil.Response(c, fiber.StatusNotFound, "Submitter not found", nil)
	}

	ok, err := h.completedDoc.IsSubmissionFullyCompleted(c.Context(), submissionID)
	if err != nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to check completion", nil)
	}
	if !ok {
		return webutil.Response(c, fiber.StatusConflict, "Submission not completed yet", nil)
	}

	// Ensure we have an absolute base URL stored for QR codes.
	baseURL := fmt.Sprintf("%s://%s", c.Protocol(), c.Get("Host"))
	_, _ = h.pool.Exec(c.Context(), `
		UPDATE submission
		SET preferences = jsonb_set(COALESCE(preferences, '{}'::jsonb), '{public_base_url}', to_jsonb($2::text), true),
		    updated_at = NOW()
		WHERE id = $1
		  AND COALESCE(preferences->>'public_base_url', '') = ''
	`, submissionID, baseURL)

	path, err := h.completedDoc.EnsureCertificatePDF(c.Context(), submissionID)
	if err != nil {
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to build certificate", map[string]any{"error": err.Error()})
	}

	return c.Download(path, fmt.Sprintf("submission_%s_certificate.pdf", submissionID))
}

func (h *PublicSigningHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/sign/:slug", h.GetBySlug)
	router.Post("/sign/:slug/open", h.Open)
	router.Post("/sign/:slug/update", h.UpdateSubmitter)
	router.Post("/sign/:slug/complete", h.Complete)
	router.Get("/sign/:slug/document", h.GetCompletedDocument)
	router.Get("/sign/:slug/certificate", h.GetCertificate)
	router.Post("/sign/:slug/decline", h.Decline)
}

func (h *PublicSigningHandler) finalizeIfCompleted(ctx context.Context, submissionID string, baseURL string) {
	if h.pool == nil || h.completedDoc == nil {
		return
	}

	// Atomically mark "completed_notified_at" only once, and only when all submitters completed.
	var marked string
	err := h.pool.QueryRow(ctx, `
		WITH all_done AS (
			SELECT bool_and(COALESCE(status, 'pending') = 'completed') AS ok
			FROM submitter
			WHERE submission_id = $1
		), upd AS (
			UPDATE submission
			SET preferences = jsonb_set(
				jsonb_set(
					jsonb_set(COALESCE(preferences, '{}'::jsonb), '{completed_at}', to_jsonb(NOW()), true),
					'{completed_notified_at}', to_jsonb(NOW()), true
				),
				'{public_base_url}', to_jsonb($2::text), true
			),
			    updated_at = NOW()
			WHERE id = $1
			  AND NOT (COALESCE(preferences, '{}'::jsonb) ? 'completed_notified_at')
			  AND (SELECT ok FROM all_done)
			RETURNING id::text
		)
		SELECT id FROM upd
	`, submissionID, strings.TrimRight(baseURL, "/")).Scan(&marked)
	if err != nil || marked == "" {
		return // not completed yet OR already sent
	}

	// Ensure the completed PDF exists (cached).
	_, _ = h.completedDoc.EnsureCompletedPDF(ctx, submissionID)

	// Send notifications (best-effort) to all submitters with provided contact info.
	rows, err := h.pool.Query(ctx, `
		SELECT COALESCE(email, ''), COALESCE(phone, ''), slug
		FROM submitter
		WHERE submission_id = $1
	`, submissionID)
	if err != nil {
		return
	}
	defer rows.Close()

	baseURL = strings.TrimRight(baseURL, "/")
	for rows.Next() {
		var email, phone, slug string
		if err := rows.Scan(&email, &phone, &slug); err != nil {
			continue
		}
		downloadURL := fmt.Sprintf("%s/public/sign/%s/document", baseURL, slug)

		if strings.TrimSpace(email) != "" && h.notificationSvc != nil && h.notificationSvc.CanSend(models.NotificationTypeEmail) {
			n := &models.Notification{
				ID:        uuid.NewString(),
				Type:      models.NotificationTypeEmail,
				Recipient: email,
				Subject:   "Completed document",
				Body:      fmt.Sprintf("The completed document is ready.\n\nDownload: %s\n", downloadURL),
				Context:   map[string]any{},
			}
			if err := h.notificationSvc.Send(n); err != nil {
				log.Warn().Err(err).Str("email", email).Msg("Failed to send completion email")
			}
		}

		if strings.TrimSpace(phone) != "" && h.notificationSvc != nil && h.notificationSvc.CanSend(models.NotificationTypeSMS) {
			n := &models.Notification{
				ID:        uuid.NewString(),
				Type:      models.NotificationTypeSMS,
				Recipient: phone,
				Body:      fmt.Sprintf("Your document is completed. Download: %s", downloadURL),
				Context:   map[string]any{},
			}
			if err := h.notificationSvc.Send(n); err != nil {
				log.Warn().Err(err).Str("phone", phone).Msg("Failed to send completion SMS")
			}
		}
	}
}

// getClientIP extracts the real client IP address from the request
// It checks X-Forwarded-For, X-Real-IP headers first, then falls back to c.IP()
func getClientIP(c *fiber.Ctx) string {
	// Check X-Forwarded-For header (first IP in the chain)
	forwardedFor := c.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	realIP := c.Get("X-Real-IP")
	if realIP != "" {
		return strings.TrimSpace(realIP)
	}

	// Fallback to Fiber's IP() method
	return c.IP()
}

// NOTE: Do not add helpers that discard context cancel funcs; always call cancel().

