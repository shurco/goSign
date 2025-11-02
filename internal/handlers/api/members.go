package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// MemberHandler handles organization member-related requests
type MemberHandler struct {
	organizationQueries *queries.OrganizationQueries
}

// NewMemberHandler creates a new member handler
func NewMemberHandler(organizationQueries *queries.OrganizationQueries) *MemberHandler {
	return &MemberHandler{
		organizationQueries: organizationQueries,
	}
}

// GetOrganizationMembers returns all members of an organization
// @Summary Get organization members
// @Description Get all members of an organization
// @Tags organizations
// @Produce json
// @Param organization_id path string true "Organization ID"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/organizations/{organization_id}/members [get]
func (h *MemberHandler) GetOrganizationMembers(c *fiber.Ctx) error {
	orgID := c.Params("organization_id")
	if orgID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Organization ID is required", nil)
	}

	// Get members
	members, err := h.organizationQueries.GetOrganizationMembers(c.Context(), orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get organization members")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get members", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Members retrieved successfully", map[string]interface{}{
		"members": members,
	})
}

// UpdateMemberRole updates a member's role in an organization
// @Summary Update member role
// @Description Update a member's role (admin/owner permissions required)
// @Tags organizations
// @Accept json
// @Produce json
// @Param organization_id path string true "Organization ID"
// @Param member_id path string true "Member ID"
// @Param request body map[string]string true "Role update data"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/organizations/{organization_id}/members/{member_id}/role [put]
func (h *MemberHandler) UpdateMemberRole(c *fiber.Ctx) error {
	orgID := c.Params("organization_id")
	memberID := c.Params("member_id")

	if orgID == "" || memberID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Organization ID and Member ID are required", nil)
	}

	// Parse request body
	var req map[string]string
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse role update request")
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	roleStr, exists := req["role"]
	if !exists {
		return webutil.Response(c, fiber.StatusBadRequest, "Role is required", nil)
	}

	// Validate role
	var newRole models.OrganizationRole
	switch roleStr {
	case string(models.OrganizationRoleOwner):
		newRole = models.OrganizationRoleOwner
	case string(models.OrganizationRoleAdmin):
		newRole = models.OrganizationRoleAdmin
	case string(models.OrganizationRoleMember):
		newRole = models.OrganizationRoleMember
	case string(models.OrganizationRoleViewer):
		newRole = models.OrganizationRoleViewer
	default:
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid role", nil)
	}

	// Get current user permissions
	userIDStr, err := GetUserID(c)
	if err != nil {
		return err
	}

	// Check if user has permission to update roles
	userMember, err := h.organizationQueries.GetOrganizationMember(c.Context(), orgID, userIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check user membership")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to check permissions", nil)
	}

	if userMember == nil {
		return webutil.Response(c, fiber.StatusForbidden, "Access denied", nil)
	}

	// Check role hierarchy permissions
	canUpdateRole := false
	switch userMember.Role {
	case models.OrganizationRoleOwner:
		// Owner can update any role
		canUpdateRole = true
	case models.OrganizationRoleAdmin:
		// Admin can update member/viewer roles, but not owner/admin
		canUpdateRole = newRole == models.OrganizationRoleMember || newRole == models.OrganizationRoleViewer
	default:
		canUpdateRole = false
	}

	if !canUpdateRole {
		return webutil.Response(c, fiber.StatusForbidden, "Insufficient permissions to update this role", nil)
	}

	// Get member to update
	members, err := h.organizationQueries.GetOrganizationMembers(c.Context(), orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get organization members")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get members", nil)
	}

	var targetMember *models.OrganizationMember
	for _, member := range members {
		if member.ID == memberID {
			targetMember = &member
			break
		}
	}

	if targetMember == nil {
		return webutil.Response(c, fiber.StatusNotFound, "Member not found", nil)
	}

	// Cannot change owner's role
	if targetMember.Role == models.OrganizationRoleOwner {
		return webutil.Response(c, fiber.StatusForbidden, "Cannot change organization owner's role", nil)
	}

	// Update member role
	if err := h.organizationQueries.UpdateOrganizationMember(c.Context(), orgID, targetMember.UserID, newRole); err != nil {
		log.Error().Err(err).Msg("Failed to update member role")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to update member role", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Member role updated successfully", nil)
}

// RemoveOrganizationMember removes a member from an organization
// @Summary Remove member
// @Description Remove a member from organization (admin/owner permissions required)
// @Tags organizations
// @Produce json
// @Param organization_id path string true "Organization ID"
// @Param member_id path string true "Member ID"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/organizations/{organization_id}/members/{member_id} [delete]
func (h *MemberHandler) RemoveOrganizationMember(c *fiber.Ctx) error {
	orgID := c.Params("organization_id")
	memberID := c.Params("member_id")

	if orgID == "" || memberID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Organization ID and Member ID are required", nil)
	}

	// Get current user permissions
	userIDStr, err := GetUserID(c)
	if err != nil {
		return err
	}

	// Check if user has permission to remove members
	userMember, err := h.organizationQueries.GetOrganizationMember(c.Context(), orgID, userIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check user membership")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to check permissions", nil)
	}

	if userMember == nil {
		return webutil.Response(c, fiber.StatusForbidden, "Access denied", nil)
	}

	// Get all members to find target member
	members, err := h.organizationQueries.GetOrganizationMembers(c.Context(), orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get organization members")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get members", nil)
	}

	var targetMember *models.OrganizationMember
	for _, member := range members {
		if member.ID == memberID {
			targetMember = &member
			break
		}
	}

	if targetMember == nil {
		return webutil.Response(c, fiber.StatusNotFound, "Member not found", nil)
	}

	// Cannot remove owner
	if targetMember.Role == models.OrganizationRoleOwner {
		return webutil.Response(c, fiber.StatusForbidden, "Cannot remove organization owner", nil)
	}

	// Cannot remove self
	if targetMember.UserID == userIDStr {
		return webutil.Response(c, fiber.StatusForbidden, "Cannot remove yourself from organization", nil)
	}

	// Check permissions based on role hierarchy
	canRemoveMember := false
	switch userMember.Role {
	case models.OrganizationRoleOwner:
		// Owner can remove anyone except other owners
		canRemoveMember = targetMember.Role != models.OrganizationRoleOwner
	case models.OrganizationRoleAdmin:
		// Admin can remove members and viewers
		canRemoveMember = targetMember.Role == models.OrganizationRoleMember || targetMember.Role == models.OrganizationRoleViewer
	default:
		canRemoveMember = false
	}

	if !canRemoveMember {
		return webutil.Response(c, fiber.StatusForbidden, "Insufficient permissions to remove this member", nil)
	}

	// Remove member
	if err := h.organizationQueries.RemoveOrganizationMember(c.Context(), orgID, targetMember.UserID); err != nil {
		log.Error().Err(err).Msg("Failed to remove organization member")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to remove member", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Member removed successfully", nil)
}

// InviteMember sends an invitation to join the organization
// @Summary Invite member
// @Description Send invitation to join organization
// @Tags organizations
// @Accept json
// @Produce json
// @Param organization_id path string true "Organization ID"
// @Param request body map[string]string true "Invitation data"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/organizations/{organization_id}/members/invite [post]
func (h *MemberHandler) InviteMember(c *fiber.Ctx) error {
	// Get organization details for email template
	orgID := c.Params("organization_id")
	if orgID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Organization ID is required", nil)
	}

	// Get organization info for email
	org, err := h.organizationQueries.GetOrganization(c.Context(), orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get organization")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get organization", nil)
	}

	if org == nil {
		return webutil.Response(c, fiber.StatusNotFound, "Organization not found", nil)
	}

	// Parse request body
	var req map[string]string
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse invitation request")
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	email, hasEmail := req["email"]
	roleStr, hasRole := req["role"]

	if !hasEmail {
		return webutil.Response(c, fiber.StatusBadRequest, "Email is required", nil)
	}

	if !hasRole {
		return webutil.Response(c, fiber.StatusBadRequest, "Role is required", nil)
	}

	// Validate email (basic validation)
	if len(email) < 3 || len(email) > 255 {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid email format", nil)
	}

	// Validate role
	var role models.OrganizationRole
	switch roleStr {
	case string(models.OrganizationRoleAdmin):
		role = models.OrganizationRoleAdmin
	case string(models.OrganizationRoleMember):
		role = models.OrganizationRoleMember
	case string(models.OrganizationRoleViewer):
		role = models.OrganizationRoleViewer
	default:
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid role", nil)
	}

	// Get current user permissions
	userIDStr, err := GetUserID(c)
	if err != nil {
		return err
	}

	// Check if user has permission to invite members
	userMember, err := h.organizationQueries.GetOrganizationMember(c.Context(), orgID, userIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check user membership")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to check permissions", nil)
	}

	if userMember == nil {
		return webutil.Response(c, fiber.StatusForbidden, "Access denied", nil)
	}

	// Check role permissions for inviting
	canInviteRole := false
	switch userMember.Role {
	case models.OrganizationRoleOwner:
		// Owner can invite anyone
		canInviteRole = true
	case models.OrganizationRoleAdmin:
		// Admin can invite members and viewers
		canInviteRole = role == models.OrganizationRoleMember || role == models.OrganizationRoleViewer
	default:
		canInviteRole = false
	}

	if !canInviteRole {
		return webutil.Response(c, fiber.StatusForbidden, "Insufficient permissions to invite with this role", nil)
	}

	// TODO: Check if user is already a member (requires user lookup by email)
	// TODO: Check if email is already invited
	// For now, we'll allow invitations even if user exists - they can accept or decline

	// Create invitation
	invitation := &models.OrganizationInvitation{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		Email:          email,
		Role:           role,
		Token:          uuid.New().String(), // In production, use crypto/rand
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 7 days
		InvitedByID:    userIDStr,
	}

	if err := h.organizationQueries.CreateOrganizationInvitation(c.Context(), invitation); err != nil {
		log.Error().Err(err).Msg("Failed to create invitation")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create invitation", nil)
	}

	// TODO: Send email invitation
	// For now, log the invitation details (email sending will be implemented when notification service is set up)
	log.Info().
		Str("organization_id", orgID).
		Str("email", email).
		Str("role", string(role)).
		Str("invitation_token", invitation.Token).
		Msg("Organization invitation created - email sending not yet implemented")

	// For now, just return success with token (for testing)

	return webutil.Response(c, fiber.StatusCreated, "Invitation sent successfully", map[string]interface{}{
		"invitation": map[string]interface{}{
			"id":         invitation.ID,
			"email":      invitation.Email,
			"role":       invitation.Role,
			"expires_at": invitation.ExpiresAt,
		},
	})
}

// GetOrganizationInvitations returns pending invitations for an organization
// @Summary Get invitations
// @Description Get pending invitations for organization
// @Tags organizations
// @Produce json
// @Param organization_id path string true "Organization ID"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/organizations/{organization_id}/invitations [get]
func (h *MemberHandler) GetOrganizationInvitations(c *fiber.Ctx) error {
	orgID := c.Params("organization_id")
	if orgID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Organization ID is required", nil)
	}

	// Get invitations
	invitations, err := h.organizationQueries.GetOrganizationInvitations(c.Context(), orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get organization invitations")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get invitations", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Invitations retrieved successfully", map[string]interface{}{
		"invitations": invitations,
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
func (h *MemberHandler) RevokeInvitation(c *fiber.Ctx) error {
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

// RegisterRoutes registers all member routes
func (h *MemberHandler) RegisterRoutes(router fiber.Router) {
	// Member management
	router.Get("/:organization_id/members", h.GetOrganizationMembers)
	router.Put("/:organization_id/members/:member_id/role", h.UpdateMemberRole)
	router.Delete("/:organization_id/members/:member_id", h.RemoveOrganizationMember)

	// Invitations
	router.Post("/:organization_id/members/invite", h.InviteMember)
	router.Get("/:organization_id/invitations", h.GetOrganizationInvitations)
	router.Delete("/:organization_id/invitations/:invitation_id", h.RevokeInvitation)
}
