package middleware

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/pkg/utils/webutil"
)

// OrganizationPermission represents a permission that can be checked
type OrganizationPermission string

const (
	// Organization-level permissions
	OrgPermViewMembers    OrganizationPermission = "org:view_members"
	OrgPermInviteMembers  OrganizationPermission = "org:invite_members"
	OrgPermRemoveMembers  OrganizationPermission = "org:remove_members"
	OrgPermUpdateSettings OrganizationPermission = "org:update_settings"
	OrgPermDeleteOrg      OrganizationPermission = "org:delete"

	// Template permissions (within organization)
	OrgPermViewTemplates    OrganizationPermission = "templates:view"
	OrgPermCreateTemplates  OrganizationPermission = "templates:create"
	OrgPermEditTemplates    OrganizationPermission = "templates:edit"
	OrgPermDeleteTemplates  OrganizationPermission = "templates:delete"
	OrgPermExportTemplates  OrganizationPermission = "templates:export"

	// Submission permissions (within organization)
	OrgPermViewSubmissions   OrganizationPermission = "submissions:view"
	OrgPermCreateSubmissions OrganizationPermission = "submissions:create"
	OrgPermDeleteSubmissions OrganizationPermission = "submissions:delete"

	// Billing permissions
	OrgPermViewBilling   OrganizationPermission = "billing:view"
	OrgPermUpdateBilling OrganizationPermission = "billing:update"
)

// rolePermissions maps organization roles to their permissions
var rolePermissions = map[models.OrganizationRole][]OrganizationPermission{
	models.OrganizationRoleOwner: {
		// All permissions for owner
		OrgPermViewMembers, OrgPermInviteMembers, OrgPermRemoveMembers,
		OrgPermUpdateSettings, OrgPermDeleteOrg,
		OrgPermViewTemplates, OrgPermCreateTemplates, OrgPermEditTemplates,
		OrgPermDeleteTemplates, OrgPermExportTemplates,
		OrgPermViewSubmissions, OrgPermCreateSubmissions, OrgPermDeleteSubmissions,
		OrgPermViewBilling, OrgPermUpdateBilling,
	},
	models.OrganizationRoleAdmin: {
		// Administrative permissions but can't delete org
		OrgPermViewMembers, OrgPermInviteMembers, OrgPermRemoveMembers,
		OrgPermUpdateSettings,
		OrgPermViewTemplates, OrgPermCreateTemplates, OrgPermEditTemplates,
		OrgPermDeleteTemplates, OrgPermExportTemplates,
		OrgPermViewSubmissions, OrgPermCreateSubmissions, OrgPermDeleteSubmissions,
		OrgPermViewBilling, OrgPermUpdateBilling,
	},
	models.OrganizationRoleMember: {
		// Standard member permissions
		OrgPermViewMembers,
		OrgPermViewTemplates, OrgPermCreateTemplates, OrgPermEditTemplates,
		OrgPermExportTemplates,
		OrgPermViewSubmissions, OrgPermCreateSubmissions,
	},
	models.OrganizationRoleViewer: {
		// Read-only permissions
		OrgPermViewMembers,
		OrgPermViewTemplates,
		OrgPermViewSubmissions,
	},
}

// OrganizationContext represents organization context in request
type OrganizationContext struct {
	OrganizationID string
	UserID         string
	Role           models.OrganizationRole
	Permissions    []OrganizationPermission
}

// OrgPermissionService handles organization permission checks
type OrgPermissionService struct {
	// This would normally inject a repository, but for now we'll use interface
	getMemberFunc func(ctx context.Context, orgID, userID string) (*models.OrganizationMember, error)
}

// NewOrgPermissionService creates a new organization permission service
func NewOrgPermissionService(getMemberFunc func(ctx context.Context, orgID, userID string) (*models.OrganizationMember, error)) *OrgPermissionService {
	return &OrgPermissionService{
		getMemberFunc: getMemberFunc,
	}
}

// CheckOrgPermission checks if a user has a specific permission in an organization
func (s *OrgPermissionService) CheckOrgPermission(ctx context.Context, orgID, userID string, permission OrganizationPermission) (bool, error) {
	member, err := s.getMemberFunc(ctx, orgID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get member: %w", err)
	}

	if member == nil {
		return false, nil // User is not a member of this organization
	}

	permissions, exists := rolePermissions[member.Role]
	if !exists {
		return false, nil // Unknown role
	}

	// Check if the permission is in the role's permissions
	for _, p := range permissions {
		if p == permission {
			return true, nil
		}
	}

	return false, nil
}

// GetOrgPermissions returns all permissions for a user's role in an organization
func (s *OrgPermissionService) GetOrgPermissions(ctx context.Context, orgID, userID string) ([]OrganizationPermission, error) {
	member, err := s.getMemberFunc(ctx, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	if member == nil {
		return []OrganizationPermission{}, nil
	}

	permissions, exists := rolePermissions[member.Role]
	if !exists {
		return []OrganizationPermission{}, nil
	}

	return permissions, nil
}

// extractOrgUserContext extracts organization and user IDs from request context
func extractOrgUserContext(c *fiber.Ctx) (orgID, userID string, err error) {
	// Get organization ID from context
	orgIDVal := c.Locals("organization_id")
	if orgIDVal == nil {
		return "", "", fiber.NewError(fiber.StatusUnauthorized, "No organization context")
	}

	orgID, ok := orgIDVal.(string)
	if !ok {
		return "", "", fiber.NewError(fiber.StatusInternalServerError, "Invalid organization context")
	}

	// Get user ID from context
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return "", "", fiber.NewError(fiber.StatusUnauthorized, "User not authenticated")
	}

	userID, ok = userIDVal.(string)
	if !ok {
		return "", "", fiber.NewError(fiber.StatusInternalServerError, "Invalid user context")
	}

	return orgID, userID, nil
}

// RequireOrgPermission creates middleware that requires a specific permission
func RequireOrgPermission(service *OrgPermissionService, permission OrganizationPermission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		orgID, userID, err := extractOrgUserContext(c)
		if err != nil {
			return err
		}

		// Check permission
		hasPermission, err := service.CheckOrgPermission(c.Context(), orgID, userID, permission)
		if err != nil {
			return webutil.Response(c, fiber.StatusInternalServerError, "Permission check failed", nil)
		}

		if !hasPermission {
			return webutil.Response(c, fiber.StatusForbidden, "Insufficient permissions", nil)
		}

		return c.Next()
	}
}

// RequireOrgRole creates middleware that requires a minimum role level
func RequireOrgRole(service *OrgPermissionService, requiredRole models.OrganizationRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		orgID, userID, err := extractOrgUserContext(c)
		if err != nil {
			return err
		}

		// Get member info
		member, err := service.getMemberFunc(c.Context(), orgID, userID)
		if err != nil {
			return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get member info", nil)
		}

		if member == nil {
			return webutil.Response(c, fiber.StatusForbidden, "User is not a member of this organization", nil)
		}

		// Check role hierarchy: owner > admin > member > viewer
		roleHierarchy := map[models.OrganizationRole]int{
			models.OrganizationRoleViewer: 1,
			models.OrganizationRoleMember: 2,
			models.OrganizationRoleAdmin:  3,
			models.OrganizationRoleOwner:  4,
		}

		userLevel, userExists := roleHierarchy[member.Role]
		requiredLevel, requiredExists := roleHierarchy[requiredRole]

		if !userExists || !requiredExists {
			return webutil.Response(c, fiber.StatusInternalServerError, "Invalid role configuration", nil)
		}

		if userLevel < requiredLevel {
			return webutil.Response(c, fiber.StatusForbidden, "Insufficient role level", nil)
		}

		// Store role context for later use
		c.Locals("user_role", member.Role)
		c.Locals("member_id", member.ID)

		return c.Next()
	}
}

// SetOrganizationContext creates middleware that sets organization context
func SetOrganizationContext(service *OrgPermissionService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get organization ID from URL parameter or header
		orgID := c.Params("organization_id")
		if orgID == "" {
			orgID = c.Get("X-Organization-ID")
		}
		if orgID == "" {
			// Try to get from query parameter
			orgID = c.Query("organization_id")
		}

		if orgID == "" {
			return webutil.Response(c, fiber.StatusBadRequest, "Organization ID is required", nil)
		}

		// Get user ID from context
		userID := c.Locals("user_id")
		if userID == nil {
			return webutil.Response(c, fiber.StatusUnauthorized, "User not authenticated", nil)
		}

		userIDStr, ok := userID.(string)
		if !ok {
			return webutil.Response(c, fiber.StatusInternalServerError, "Invalid user context", nil)
		}

		// Get member info to validate membership and get role
		member, err := service.getMemberFunc(c.Context(), orgID, userIDStr)
		if err != nil {
			return webutil.Response(c, fiber.StatusInternalServerError, "Failed to validate organization membership", nil)
		}

		if member == nil {
			return webutil.Response(c, fiber.StatusForbidden, "User is not a member of this organization", nil)
		}

		// Set organization context
		c.Locals("organization_id", orgID)
		c.Locals("user_role", member.Role)
		c.Locals("member_id", member.ID)

		// Get permissions for this role
		permissions, err := service.GetOrgPermissions(c.Context(), orgID, userIDStr)
		if err != nil {
			return webutil.Response(c, fiber.StatusInternalServerError, "Failed to get permissions", nil)
		}

		c.Locals("user_permissions", permissions)

		return c.Next()
	}
}

// Helper functions for permission checking in handlers

// HasPermission checks if current request has the required permission
func HasPermission(c *fiber.Ctx, permission OrganizationPermission) bool {
	permissions := c.Locals("user_permissions")
	if permissions == nil {
		return false
	}

	perms, ok := permissions.([]OrganizationPermission)
	if !ok {
		return false
	}

	for _, p := range perms {
		if p == permission {
			return true
		}
	}

	return false
}

// GetCurrentOrgID returns the current organization ID from context
func GetCurrentOrgID(c *fiber.Ctx) (string, bool) {
	orgID := c.Locals("organization_id")
	if orgID == nil {
		return "", false
	}

	orgIDStr, ok := orgID.(string)
	return orgIDStr, ok
}

// GetCurrentUserRole returns the current user's role from context
func GetCurrentUserRole(c *fiber.Ctx) (models.OrganizationRole, bool) {
	role := c.Locals("user_role")
	if role == nil {
		return "", false
	}

	roleStr, ok := role.(models.OrganizationRole)
	return roleStr, ok
}
