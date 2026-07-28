package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// InvitationHandler handles organization invitation-related requests
type InvitationHandler struct {
	organizationQueries *queries.OrganizationQueries
}

// NewInvitationHandler creates a new invitation handler
func NewInvitationHandler(organizationQueries *queries.OrganizationQueries) *InvitationHandler {
	return &InvitationHandler{
		organizationQueries: organizationQueries,
	}
}

// AcceptInvitation accepts an organization invitation
// @Summary Accept invitation
// @Description Accept invitation to join organization using token
// @Tags organizations
// @Accept json
// @Produce json
// @Param token path string true "Invitation token"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 409 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /v1/invitations/{token}/accept [post]
func (h *InvitationHandler) AcceptInvitation(c fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Invitation token is required", nil)
	}

	// Get user ID from context
	userIDStr, err := GetUserID(c)
	if err != nil {
		return err
	}

	// Get invitation by token
	invitation, err := h.organizationQueries.GetOrganizationInvitation(c.Context(), token)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get invitation")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to process invitation", nil)
	}

	if invitation == nil {
		return webutil.Response(c, fiber.StatusNotFound, "Invitation not found or expired", nil)
	}

	// Check if invitation is already accepted
	if invitation.AcceptedAt != nil {
		return webutil.Response(c, fiber.StatusConflict, "Invitation already accepted", nil)
	}

	// Get organization details
	org, err := h.organizationQueries.GetOrganization(c.Context(), invitation.OrganizationID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get organization")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to process invitation", nil)
	}

	if org == nil {
		return webutil.Response(c, fiber.StatusNotFound, "Organization not found", nil)
	}

	// Check if user is already a member
	existingMember, err := h.organizationQueries.GetOrganizationMember(c.Context(), invitation.OrganizationID, userIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check membership")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to process invitation", nil)
	}

	if existingMember != nil {
		return webutil.Response(c, fiber.StatusConflict, "You are already a member of this organization", nil)
	}

	// Accept the invitation
	if err := h.organizationQueries.AcceptOrganizationInvitation(c.Context(), token, userIDStr); err != nil {
		log.Error().Err(err).Msg("Failed to accept invitation")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to accept invitation", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Invitation accepted successfully", map[string]any{
		"organization": map[string]any{
			"id":   org.ID,
			"name": org.Name,
		},
		"role": invitation.Role,
	})
}

// GetInvitationDetails gets invitation details without accepting
// @Summary Get invitation details
// @Description Get invitation details using token (for preview before accepting)
// @Tags organizations
// @Produce json
// @Param token path string true "Invitation token"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /v1/invitations/{token} [get]
func (h *InvitationHandler) GetInvitationDetails(c fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Invitation token is required", nil)
	}

	// Get invitation by token
	invitation, err := h.organizationQueries.GetOrganizationInvitation(c.Context(), token)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get invitation")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get invitation details", nil)
	}

	if invitation == nil {
		return webutil.Response(c, fiber.StatusNotFound, "Invitation not found or expired", nil)
	}

	// Get organization details
	org, err := h.organizationQueries.GetOrganization(c.Context(), invitation.OrganizationID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get organization")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get organization details", nil)
	}

	if org == nil {
		return webutil.Response(c, fiber.StatusNotFound, "Organization not found", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Invitation details retrieved", map[string]any{
		"invitation": map[string]any{
			"email":      invitation.Email,
			"role":       invitation.Role,
			"expires_at": invitation.ExpiresAt,
			"invited_by": invitation.InvitedByID,
		},
		"organization": map[string]any{
			"id":          org.ID,
			"name":        org.Name,
			"description": org.Description,
		},
	})
}

// RegisterRoutes registers all invitation routes
func (h *InvitationHandler) RegisterRoutes(router fiber.Router) {
	// Public routes (no auth required)
	router.Get("/:token", h.GetInvitationDetails)
	router.Post("/:token/accept", h.AcceptInvitation)
}
