package api

import (
	"encoding/base64"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/services/submission"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// SubmitterHandler handles requests to submitters
type SubmitterHandler struct {
	*ResourceHandler[models.Submitter] // embed generic CRUD
	submissionService                  *submission.Service
	submitterRepo                      ResourceRepository[models.Submitter]
}

// NewSubmitterHandler creates new handler
func NewSubmitterHandler(repo ResourceRepository[models.Submitter], submissionService *submission.Service) *SubmitterHandler {
	return &SubmitterHandler{
		ResourceHandler:   NewResourceHandler("submitter", repo),
		submissionService: submissionService,
		submitterRepo:     repo,
	}
}

// ResendRequest request body for resending
type ResendRequest struct {
	SubmitterID string `json:"submitter_id" validate:"required"`
}

// Resend resends invitation to submitter
// @Summary Resend invitation
// @Description Resends invitation email to a submitter
// @Tags submitters
// @Accept json
// @Produce json
// @Param body body ResendRequest true "Resend request"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Router /api/submitters/resend [post]
func (h *SubmitterHandler) Resend(c *fiber.Ctx) error {
	var req ResendRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.StatusBadRequest(c, "Invalid request body")
	}

	if req.SubmitterID == "" {
		return webutil.StatusBadRequest(c, "submitter_id is required")
	}

	// Get submitter
	submitter, err := h.submitterRepo.Get(req.SubmitterID)
	if err != nil {
		log.Error().Err(err).Str("submitter_id", req.SubmitterID).Msg("Failed to get submitter")
		return webutil.Response(c, fiber.StatusNotFound, "Submitter not found", nil)
	}

	// Check if already completed or declined
	if submitter.Status == models.SubmitterStatusCompleted {
		return webutil.Response(c, fiber.StatusBadRequest, "Cannot resend to completed submitter", nil)
	}
	if submitter.Status == models.SubmitterStatusDeclined {
		return webutil.Response(c, fiber.StatusBadRequest, "Cannot resend to declined submitter", nil)
	}

	// Resend invitation via submission service
	if err := h.submissionService.ResendInvitation(c.Context(), req.SubmitterID); err != nil {
		log.Error().Err(err).Str("submitter_id", req.SubmitterID).Msg("Failed to resend invitation")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to resend invitation", nil)
	}

	log.Info().Str("submitter_id", req.SubmitterID).Msg("Invitation resent")

	return webutil.Response(c, fiber.StatusOK, "invitation_resent", map[string]any{
		"submitter_id": req.SubmitterID,
		"status":       "resent",
	})
}

// DeclineRequest request body for declining
type DeclineRequest struct {
	SubmitterID string `json:"submitter_id" validate:"required"`
	Reason      string `json:"reason,omitempty"`
}

// Decline declines signing
// @Summary Decline signing
// @Description Marks a submitter as declined with optional reason
// @Tags submitters
// @Accept json
// @Produce json
// @Param body body DeclineRequest true "Decline request"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Router /api/submitters/decline [post]
func (h *SubmitterHandler) Decline(c *fiber.Ctx) error {
	var req DeclineRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.StatusBadRequest(c, "Invalid request body")
	}

	if req.SubmitterID == "" {
		return webutil.StatusBadRequest(c, "submitter_id is required")
	}

	// Get submitter to get submission ID
	submitter, err := h.submitterRepo.Get(req.SubmitterID)
	if err != nil {
		log.Error().Err(err).Str("submitter_id", req.SubmitterID).Msg("Failed to get submitter")
		return webutil.Response(c, fiber.StatusNotFound, "Submitter not found", nil)
	}

	// Decline submitter
	if err := h.submissionService.Decline(c.Context(), req.SubmitterID, req.Reason); err != nil {
		log.Error().Err(err).Str("submitter_id", req.SubmitterID).Msg("Failed to decline submitter")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to decline submitter", nil)
	}

	// Handle decline at submission level
	_ = h.submissionService.HandleDecline(c.Context(), submitter.SubmissionID, req.Reason)

	return webutil.Response(c, fiber.StatusOK, "submitter_declined", map[string]any{
		"submitter_id": req.SubmitterID,
		"status":       "declined",
		"reason":       req.Reason,
	})
}

// CompleteRequest request body for completing signing
type CompleteRequest struct {
	SubmitterID string                 `json:"submitter_id" validate:"required"`
	Fields      map[string]any `json:"fields" validate:"required"`
	Signature   SignatureData          `json:"signature,omitempty"`
}

// SignatureData signature data
type SignatureData struct {
	ImageBase64 string  `json:"image_base64"` // Base64 encoded signature image
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Width       float64 `json:"width"`
	Height      float64 `json:"height"`
	Page        int     `json:"page"`
}

// Complete completes signing for submitter
// @Summary Complete signing
// @Description Marks a submitter as completed and saves their data
// @Tags submitters
// @Accept json
// @Produce json
// @Param body body CompleteRequest true "Complete request"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Router /api/submitters/complete [post]
func (h *SubmitterHandler) Complete(c *fiber.Ctx) error {
	var req CompleteRequest
	if err := c.BodyParser(&req); err != nil {
		return webutil.StatusBadRequest(c, "Invalid request body")
	}

	if req.SubmitterID == "" {
		return webutil.StatusBadRequest(c, "submitter_id is required")
	}

	// Get submitter
	submitter, err := h.submitterRepo.Get(req.SubmitterID)
	if err != nil {
		log.Error().Err(err).Str("submitter_id", req.SubmitterID).Msg("Failed to get submitter")
		return webutil.Response(c, fiber.StatusNotFound, "Submitter not found", nil)
	}

	// Save submitter data (fields + signature)
	// Merge existing metadata with new fields
	if submitter.Metadata == nil {
		submitter.Metadata = make(map[string]any)
	}
	
	// Store fields
	if req.Fields != nil {
		submitter.Metadata["fields"] = req.Fields
	}

	// Store signature data
	if req.Signature.ImageBase64 != "" {
		// Decode base64 signature
		sigData, err := base64.StdEncoding.DecodeString(req.Signature.ImageBase64)
		if err == nil {
			submitter.Metadata["signature"] = map[string]any{
				"image_base64": req.Signature.ImageBase64,
				"image_size":   len(sigData),
				"x":            req.Signature.X,
				"y":            req.Signature.Y,
				"width":        req.Signature.Width,
				"height":       req.Signature.Height,
				"page":         req.Signature.Page,
			}
		}
	}

	// Update in database
	if err := h.submitterRepo.Update(req.SubmitterID, submitter); err != nil {
		log.Error().Err(err).Msg("Failed to update submitter")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to save data", nil)
	}

	// Complete submitter status
	if err := h.submissionService.Complete(c.Context(), req.SubmitterID); err != nil {
		log.Error().Err(err).Str("submitter_id", req.SubmitterID).Msg("Failed to complete submitter")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to complete submitter", nil)
	}

	// Check if all submitters completed
	_ = h.submissionService.CheckCompletion(c.Context(), submitter.SubmissionID)

	return webutil.Response(c, fiber.StatusOK, "submitter_completed", map[string]any{
		"submitter_id": req.SubmitterID,
		"status":       "completed",
	})
}

// RegisterRoutes registers all routes for submitters
func (h *SubmitterHandler) RegisterRoutes(router fiber.Router) {
	// Generic CRUD routes
	h.ResourceHandler.RegisterRoutes(router)

	// Specific business operations
	router.Post("/resend", h.Resend)
	router.Post("/decline", h.Decline)
	router.Post("/complete", h.Complete)
}

