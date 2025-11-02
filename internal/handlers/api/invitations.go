package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/models"
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
// @Router /api/v1/invitations/{token}/accept [post]
func (h *InvitationHandler) AcceptInvitation(c *fiber.Ctx) error {
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

	return webutil.Response(c, fiber.StatusOK, "Invitation accepted successfully", map[string]interface{}{
		"organization": map[string]interface{}{
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
// @Router /api/v1/invitations/{token} [get]
func (h *InvitationHandler) GetInvitationDetails(c *fiber.Ctx) error {
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

	return webutil.Response(c, fiber.StatusOK, "Invitation details retrieved", map[string]interface{}{
		"invitation": map[string]interface{}{
			"email":      invitation.Email,
			"role":       invitation.Role,
			"expires_at": invitation.ExpiresAt,
			"invited_by": invitation.InvitedByID,
		},
		"organization": map[string]interface{}{
			"id":          org.ID,
			"name":        org.Name,
			"description": org.Description,
		},
	})
}

// RevokeInvitation revokes an invitation (admin/owner only)
// @Summary Revoke invitation
// @Description Revoke pending invitation (admin/owner permissions required)
// @Tags organizations
// @Produce json
// @Param organization_id path string true "Organization ID"
// @Param invitation_id path string true "Invitation ID"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/organizations/{organization_id}/invitations/{invitation_id} [delete]
func (h *InvitationHandler) RevokeInvitation(c *fiber.Ctx) error {
	orgID := c.Params("organization_id")
	invitationID := c.Params("invitation_id")

	if orgID == "" || invitationID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Organization ID and Invitation ID are required", nil)
	}

	// Get current user permissions
	userID := c.Locals("user_id")
	if userID == nil {
		return webutil.Response(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return webutil.Response(c, fiber.StatusInternalServerError, "Invalid user context", nil)
	}

	// Check if user has permission to revoke invitations
	userMember, err := h.organizationQueries.GetOrganizationMember(c.Context(), orgID, userIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check user membership")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to check permissions", nil)
	}

	if userMember == nil {
		return webutil.Response(c, fiber.StatusForbidden, "Access denied", nil)
	}

	// Only admins and owners can revoke invitations
	if userMember.Role != models.OrganizationRoleAdmin && userMember.Role != models.OrganizationRoleOwner {
		return webutil.Response(c, fiber.StatusForbidden, "Insufficient permissions to revoke invitations", nil)
	}

	// Get all invitations to find the one to revoke
	invitations, err := h.organizationQueries.GetOrganizationInvitations(c.Context(), orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get organization invitations")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get invitations", nil)
	}

	var targetInvitation *models.OrganizationInvitation
	for _, inv := range invitations {
		if inv.ID == invitationID {
			targetInvitation = &inv
			break
		}
	}

	if targetInvitation == nil {
		return webutil.Response(c, fiber.StatusNotFound, "Invitation not found", nil)
	}

	// Revoke invitation (delete it)
	if err := h.organizationQueries.DeleteOrganizationInvitation(c.Context(), orgID, targetInvitation.Email); err != nil {
		log.Error().Err(err).Msg("Failed to revoke invitation")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to revoke invitation", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Invitation revoked successfully", nil)
}

// RegisterRoutes registers all invitation routes
func (h *InvitationHandler) RegisterRoutes(router fiber.Router) {
	// Public routes (no auth required)
	router.Get("/:token", h.GetInvitationDetails)
	router.Post("/:token/accept", h.AcceptInvitation)

	// Protected routes (organization context required)
	// Note: These would be registered in the organization group
}

