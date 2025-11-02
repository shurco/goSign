package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/services/submission"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// SubmissionHandler handles requests to submissions
type SubmissionHandler struct {
	*ResourceHandler[models.Submission] // embed generic CRUD
	submissionService                   *submission.Service
}

// NewSubmissionHandler creates new handler
func NewSubmissionHandler(repo ResourceRepository[models.Submission], submissionService *submission.Service) *SubmissionHandler {
	return &SubmissionHandler{
		ResourceHandler:   NewResourceHandler("submission", repo),
		submissionService: submissionService,
	}
}

// SendRequest request body for sending submission
type SendRequest struct {
	SubmissionID string `json:"submission_id" validate:"required"`
}

// Send sends invitations to submitters
// @Summary Send submission invitations
// @Description Sends invitations to all submitters and changes submission status to in_progress
// @Tags submissions
// @Accept json
// @Produce json
// @Param body body SendRequest true "Send request"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Router /api/submissions/send [post]
func (h *SubmissionHandler) Send(c *fiber.Ctx) error {
	var req SendRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if req.SubmissionID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "submission_id is required", nil)
	}

	if err := h.submissionService.Send(c.Context(), req.SubmissionID); err != nil {
		log.Error().Err(err).Str("submission_id", req.SubmissionID).Msg("Failed to send submission")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to send submission", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "submission_sent", map[string]any{
		"submission_id": req.SubmissionID,
		"status":        "sent",
	})
}

// BulkCreateRequest request body for bulk creation
type BulkCreateRequest struct {
	TemplateID   string              `json:"template_id" validate:"required"`
	SigningMode  models.SigningMode  `json:"signing_mode,omitempty"`
	Submitters   []SubmitterBulkData `json:"submitters" validate:"required,min=1"`
}

// SubmitterBulkData submitter data for bulk operation
type SubmitterBulkData struct {
	Name  string            `json:"name" validate:"required"`
	Email string            `json:"email" validate:"required,email"`
	Phone string            `json:"phone,omitempty"`
	Data  map[string]string `json:"data,omitempty"` // additional data for filling
}

// BulkCreate bulk creation submissions
// @Summary Bulk create submissions
// @Description Creates multiple submissions from a single template
// @Tags submissions
// @Accept json
// @Produce json
// @Param body body BulkCreateRequest true "Bulk create request"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Router /api/submissions/bulk [post]
func (h *SubmissionHandler) BulkCreate(c *fiber.Ctx) error {
	var req BulkCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if req.TemplateID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "template_id is required", nil)
	}

	if len(req.Submitters) == 0 {
		return webutil.Response(c, fiber.StatusBadRequest, "at least one submitter is required", nil)
	}

	// Get user ID from context
	userID, err := GetUserID(c)
	if err != nil {
		return err
	}

	// Create submissions for each submitter
	createdIDs := make([]string, 0, len(req.Submitters))
	for _, submitterData := range req.Submitters {
		// Create submission
		input := submission.CreateSubmissionInput{
			TemplateID:  req.TemplateID,
			CreatedByID: userID,
			SigningMode: req.SigningMode,
			Submitters: []submission.SubmitterInput{
				{
					Name:  submitterData.Name,
					Email: submitterData.Email,
					Phone: submitterData.Phone,
				},
			},
		}

		sub, err := h.submissionService.Create(c.Context(), input)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create submission in bulk")
			continue // continue with others
		}

		// Send immediately
		if err := h.submissionService.Send(c.Context(), sub.ID); err != nil {
			log.Error().Err(err).Str("submission_id", sub.ID).Msg("Failed to send submission in bulk")
		}

		createdIDs = append(createdIDs, sub.ID)
	}

	return webutil.Response(c, fiber.StatusOK, "bulk_created", map[string]any{
		"total_requested": len(req.Submitters),
		"total_created":   len(createdIDs),
		"submission_ids":  createdIDs,
	})
}

// ExpireRequest request body for expiring submission
type ExpireRequest struct {
	SubmissionID string `json:"submission_id" validate:"required"`
}

// Expire marks submission as expired
// @Summary Expire submission
// @Description Marks a submission as expired
// @Tags submissions
// @Accept json
// @Produce json
// @Param body body ExpireRequest true "Expire request"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Router /api/submissions/expire [post]
func (h *SubmissionHandler) Expire(c *fiber.Ctx) error {
	var req ExpireRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if req.SubmissionID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "submission_id is required", nil)
	}

	if err := h.submissionService.Expire(c.Context(), req.SubmissionID); err != nil {
		log.Error().Err(err).Str("submission_id", req.SubmissionID).Msg("Failed to expire submission")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to expire submission", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "submission_expired", map[string]any{
		"submission_id": req.SubmissionID,
		"status":        "expired",
	})
}

// RegisterRoutes registers all routes for submissions
func (h *SubmissionHandler) RegisterRoutes(router fiber.Router) {
	// Generic CRUD routes
	h.ResourceHandler.RegisterRoutes(router)

	// Specific business operations
	router.Post("/send", h.Send)
	router.Post("/bulk", h.BulkCreate)
	router.Post("/expire", h.Expire)
}

