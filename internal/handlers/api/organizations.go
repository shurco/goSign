package api

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/shurco/gosign/internal/middleware"
	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/storage/redis"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// OrganizationHandler handles organization-related requests
type OrganizationHandler struct {
	organizationQueries *queries.OrganizationQueries
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(organizationQueries *queries.OrganizationQueries) *OrganizationHandler {
	return &OrganizationHandler{
		organizationQueries: organizationQueries,
	}
}


// CreateOrganization creates a new organization
// @Summary Create organization
// @Description Create a new organization and make the current user its owner
// @Tags organizations
// @Accept json
// @Produce json
// @Param request body models.Organization true "Organization data"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/organizations [post]
func (h *OrganizationHandler) CreateOrganization(c *fiber.Ctx) error {
	userIDStr, err := GetUserID(c)
	if err != nil {
		return err
	}

	// Parse request body
	var req models.Organization
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse organization request")
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	// Validate input
	if req.Name == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Organization name is required", nil)
	}

	if len(req.Name) > 100 {
		return webutil.Response(c, fiber.StatusBadRequest, "Organization name must be less than 100 characters", nil)
	}

	if len(req.Description) > 500 {
		return webutil.Response(c, fiber.StatusBadRequest, "Organization description must be less than 500 characters", nil)
	}

	// Create organization
	org := &models.Organization{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     userIDStr,
	}

	if err := h.organizationQueries.CreateOrganization(c.Context(), org); err != nil {
		log.Error().Err(err).Msg("Failed to create organization")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create organization", nil)
	}

	// Add creator as owner member
	ownerMember := &models.OrganizationMember{
		ID:             uuid.New().String(),
		OrganizationID: org.ID,
		UserID:         userIDStr,
		Role:           models.OrganizationRoleOwner,
	}

	if err := h.organizationQueries.AddOrganizationMember(c.Context(), ownerMember); err != nil {
		log.Error().Err(err).Msg("Failed to add owner member")
		// Don't return error as organization was created successfully
	}

	return webutil.Response(c, fiber.StatusCreated, "Organization created successfully", map[string]interface{}{
		"organization": org,
	})
}

// GetUserOrganizations returns organizations where the user is a member
// @Summary Get user organizations
// @Description Get all organizations where the current user is a member
// @Tags organizations
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/organizations [get]
func (h *OrganizationHandler) GetUserOrganizations(c *fiber.Ctx) error {
	userIDStr, err := GetUserID(c)
	if err != nil {
		return err
	}

	// Get organizations
	organizations, err := h.organizationQueries.GetUserOrganizations(c.Context(), userIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user organizations")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get organizations", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Organizations retrieved successfully", map[string]interface{}{
		"organizations": organizations,
	})
}

// GetOrganization returns organization details
// @Summary Get organization
// @Description Get organization details by ID
// @Tags organizations
// @Produce json
// @Param organization_id path string true "Organization ID"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/organizations/{organization_id} [get]
func (h *OrganizationHandler) GetOrganization(c *fiber.Ctx) error {
	orgID := c.Params("organization_id")
	if orgID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Organization ID is required", nil)
	}

	// Get organization
	org, err := h.organizationQueries.GetOrganization(c.Context(), orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get organization")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get organization", nil)
	}

	if org == nil {
		return webutil.Response(c, fiber.StatusNotFound, "Organization not found", nil)
	}

	// Check if user has permission to view this organization
	userIDStr, err := GetUserID(c)
	if err == nil {
		// Check if user is a member
		member, err := h.organizationQueries.GetOrganizationMember(c.Context(), orgID, userIDStr)
		if err != nil {
			log.Error().Err(err).Msg("Failed to check organization membership")
			return webutil.Response(c, fiber.StatusInternalServerError, "Failed to check permissions", nil)
		}

		if member == nil {
			return webutil.Response(c, fiber.StatusForbidden, "Access denied", nil)
		}
	}

	return webutil.Response(c, fiber.StatusOK, "Organization retrieved successfully", map[string]interface{}{
		"organization": org,
	})
}

// UpdateOrganization updates organization details
// @Summary Update organization
// @Description Update organization name and description
// @Tags organizations
// @Accept json
// @Produce json
// @Param organization_id path string true "Organization ID"
// @Param request body map[string]string true "Update data"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/organizations/{organization_id} [put]
func (h *OrganizationHandler) UpdateOrganization(c *fiber.Ctx) error {
	orgID := c.Params("organization_id")
	if orgID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Organization ID is required", nil)
	}

	// Parse request body
	var req map[string]string
	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse update request")
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	name, hasName := req["name"]
	description, hasDescription := req["description"]

	if !hasName && !hasDescription {
		return webutil.Response(c, fiber.StatusBadRequest, "At least one field must be provided", nil)
	}

	if hasName && (name == "" || len(name) > 100) {
		return webutil.Response(c, fiber.StatusBadRequest, "Invalid organization name", nil)
	}

	if hasDescription && len(description) > 500 {
		return webutil.Response(c, fiber.StatusBadRequest, "Organization description too long", nil)
	}

	// Check if user has permission to update organization
	userIDStr, err := GetUserID(c)
	if err != nil {
		return err
	}

	// Check membership and role
	member, err := h.organizationQueries.GetOrganizationMember(c.Context(), orgID, userIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check organization membership")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to check permissions", nil)
	}

	if member == nil {
		return webutil.Response(c, fiber.StatusForbidden, "Access denied", nil)
	}

	// Only admins and owners can update organization
	if member.Role != models.OrganizationRoleAdmin && member.Role != models.OrganizationRoleOwner {
		return webutil.Response(c, fiber.StatusForbidden, "Insufficient permissions", nil)
	}

	// Update organization
	if err := h.organizationQueries.UpdateOrganization(c.Context(), orgID, name, description); err != nil {
		log.Error().Err(err).Msg("Failed to update organization")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to update organization", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Organization updated successfully", nil)
}

// DeleteOrganization deletes an organization (owner only)
// @Summary Delete organization
// @Description Delete organization (only owner can perform this action)
// @Tags organizations
// @Produce json
// @Param organization_id path string true "Organization ID"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/organizations/{organization_id} [delete]
func (h *OrganizationHandler) DeleteOrganization(c *fiber.Ctx) error {
	orgID := c.Params("organization_id")
	if orgID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Organization ID is required", nil)
	}

	// Check if user has permission to delete organization (owner only)
	userIDStr, err := GetUserID(c)
	if err != nil {
		return err
	}

	// Check membership and role
	member, err := h.organizationQueries.GetOrganizationMember(c.Context(), orgID, userIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check organization membership")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to check permissions", nil)
	}

	if member == nil {
		return webutil.Response(c, fiber.StatusForbidden, "Access denied", nil)
	}

	// Only owners can delete organization
	if member.Role != models.OrganizationRoleOwner {
		return webutil.Response(c, fiber.StatusForbidden, "Only organization owner can delete organization", nil)
	}

	// Delete organization
	if err := h.organizationQueries.DeleteOrganization(c.Context(), orgID); err != nil {
		log.Error().Err(err).Msg("Failed to delete organization")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to delete organization", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Organization deleted successfully", nil)
}

// SwitchOrganization switches the current user's context to a different organization
// @Summary Switch organization context
// @Description Switch to a different organization and get new access tokens
// @Tags organizations
// @Produce json
// @Param organization_id path string true "Organization ID"
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/v1/organizations/{organization_id}/switch [post]
func (h *OrganizationHandler) SwitchOrganization(c *fiber.Ctx) error {
	orgID := c.Params("organization_id")
	if orgID == "" {
		return webutil.Response(c, fiber.StatusBadRequest, "Organization ID is required", nil)
	}

	userIDStr, err := GetUserID(c)
	if err != nil {
		return err
	}

	// Check if user is a member of this organization
	member, err := h.organizationQueries.GetOrganizationMember(c.Context(), orgID, userIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check organization membership")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to check membership", nil)
	}

	if member == nil {
		return webutil.Response(c, fiber.StatusForbidden, "You are not a member of this organization", nil)
	}

	// Get organization details
	org, err := h.organizationQueries.GetOrganization(c.Context(), orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get organization")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get organization", nil)
	}

	if org == nil {
		return webutil.Response(c, fiber.StatusNotFound, "Organization not found", nil)
	}

	// Get user record for token creation
	userRecord, err := queries.DB.GetUserByID(c.Context(), userIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user record")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get user data", nil)
	}

	if userRecord == nil {
		return webutil.Response(c, fiber.StatusNotFound, "User not found", nil)
	}

	// Create new tokens with organization context
	// Note: We need to import the handlers package to access createAuthTokensWithOrg
	// For now, we'll create tokens directly using middleware
	modelUser := &models.User{
		ID:    userRecord.ID,
		Name:  fmt.Sprintf("%s %s", userRecord.FirstName, userRecord.LastName),
		Email: userRecord.Email,
	}

	accessToken, err := middleware.CreateTokenWithOrg(modelUser, orgID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create access token")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create access token", nil)
	}

	refreshToken, err := middleware.CreateRefreshToken(userIDStr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create refresh token")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to create refresh token", nil)
	}

	// Store refresh token in Redis
	refreshKey := fmt.Sprintf("refresh_token:%s", userIDStr)
	if err := redis.Conn.Set(refreshKey, refreshToken, 7*24*time.Hour); err != nil {
		log.Error().Err(err).Msg("Failed to store refresh token")
		return webutil.Response(c, fiber.StatusInternalServerError, "Failed to store refresh token", nil)
	}

	return webutil.Response(c, fiber.StatusOK, "Organization switched successfully", map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"organization": map[string]interface{}{
			"id":   org.ID,
			"name": org.Name,
		},
		"role": member.Role,
	})
}

// RegisterRoutes registers all organization routes
func (h *OrganizationHandler) RegisterRoutes(router fiber.Router) {
	// Organization CRUD
	router.Post("", h.CreateOrganization)
	router.Get("", h.GetUserOrganizations)
	router.Get("/:organization_id", h.GetOrganization)
	router.Put("/:organization_id", h.UpdateOrganization)
	router.Delete("/:organization_id", h.DeleteOrganization)

	// Organization switching
	router.Post("/:organization_id/switch", h.SwitchOrganization)
}
