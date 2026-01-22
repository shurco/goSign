package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shurco/gosign/internal/models"
)

// OrganizationQueries handles organization-related database operations
type OrganizationQueries struct {
	*pgxpool.Pool
}

// NewOrganizationQueries creates a new OrganizationQueries instance
func NewOrganizationQueries(pool *pgxpool.Pool) *OrganizationQueries {
	return &OrganizationQueries{
		Pool: pool,
	}
}

// CreateOrganization creates a new organization
func (q *OrganizationQueries) CreateOrganization(ctx context.Context, org *models.Organization) error {
	query := `
		INSERT INTO "organization" (
			"id", "name", "description", "owner_id", "created_at", "updated_at"
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	now := time.Now()
	if org.CreatedAt.IsZero() {
		org.CreatedAt = now
	}
	if org.UpdatedAt.IsZero() {
		org.UpdatedAt = now
	}

	_, err := q.Exec(ctx, query,
		org.ID,
		org.Name,
		org.Description,
		org.OwnerID,
		org.CreatedAt,
		org.UpdatedAt,
	)

	return err
}

// GetOrganization retrieves an organization by ID
func (q *OrganizationQueries) GetOrganization(ctx context.Context, id string) (*models.Organization, error) {
	query := `
		SELECT id, name, description, owner_id, created_at, updated_at
		FROM "organization"
		WHERE id = $1
	`

	var org models.Organization
	err := q.QueryRow(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&org.Description,
		&org.OwnerID,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &org, nil
}

// GetUserOrganizations retrieves all organizations where the user is a member
func (q *OrganizationQueries) GetUserOrganizations(ctx context.Context, userID string) ([]models.Organization, error) {
	query := `
		SELECT DISTINCT o.id, o.name, o.description, o.owner_id, o.created_at, o.updated_at
		FROM "organization" o
		INNER JOIN "organization_member" om ON o.id = om.organization_id
		WHERE om.user_id = $1
		ORDER BY o.created_at DESC
	`

	rows, err := q.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var organizations []models.Organization
	for rows.Next() {
		var org models.Organization
		err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.Description,
			&org.OwnerID,
			&org.CreatedAt,
			&org.UpdatedAt,
		)
		if err != nil {
			// Error returned to handler for logging with context
			return nil, err
		}
		organizations = append(organizations, org)
	}

	if err := rows.Err(); err != nil {
		// Error returned to handler for logging with context
		return nil, err
		
	}

	return organizations, nil
}

// UpdateOrganization updates an organization
func (q *OrganizationQueries) UpdateOrganization(ctx context.Context, id string, name, description string) error {
	query := `
		UPDATE "organization"
		SET name = $2, description = $3, updated_at = $4
		WHERE id = $1
	`

	_, err := q.Exec(ctx, query, id, name, description, time.Now())
	if err != nil {
		// Error returned to handler for logging with context
		return err
		
	}
	return nil
}

// DeleteOrganization deletes an organization (soft delete by setting deleted_at)
func (q *OrganizationQueries) DeleteOrganization(ctx context.Context, id string) error {
	query := `
		UPDATE "organization"
		SET updated_at = $2
		WHERE id = $1
	`

	_, err := q.Exec(ctx, query, id, time.Now())
	if err != nil {
		// Error returned to handler for logging with context
		return err
		
	}
	return nil
}

// AddOrganizationMember adds a member to an organization
func (q *OrganizationQueries) AddOrganizationMember(ctx context.Context, member *models.OrganizationMember) error {
	query := `
		INSERT INTO "organization_member" (
			"id", "organization_id", "user_id", "role", "joined_at", "updated_at"
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	now := time.Now()
	if member.JoinedAt.IsZero() {
		member.JoinedAt = now
	}
	if member.UpdatedAt.IsZero() {
		member.UpdatedAt = now
	}

	_, err := q.Exec(ctx, query,
		member.ID,
		member.OrganizationID,
		member.UserID,
		member.Role,
		member.JoinedAt,
		member.UpdatedAt,
	)

	if err != nil {
		// Error returned to handler for logging with context
		return err
		
	}
	return nil
}

// GetOrganizationMember retrieves a member by organization and user ID
func (q *OrganizationQueries) GetOrganizationMember(ctx context.Context, orgID, userID string) (*models.OrganizationMember, error) {
	query := `
		SELECT id, organization_id, user_id, role, joined_at, updated_at
		FROM "organization_member"
		WHERE organization_id = $1 AND user_id = $2
	`

	var member models.OrganizationMember
	err := q.QueryRow(ctx, query, orgID, userID).Scan(
		&member.ID,
		&member.OrganizationID,
		&member.UserID,
		&member.Role,
		&member.JoinedAt,
		&member.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil // Member not found
		}
		// Error returned to handler for logging with context
		return nil, err
		
	}

	return &member, nil
}

// GetOrganizationMembers retrieves all members of an organization
func (q *OrganizationQueries) GetOrganizationMembers(ctx context.Context, orgID string) ([]models.OrganizationMember, error) {
	query := `
		SELECT om.id, om.organization_id, om.user_id, om.role, om.joined_at, om.updated_at,
		       a.name as user_name, a.timezone, a.locale,
		       COALESCE(u.email, '') as user_email,
		       COALESCE(u.first_name, '') as first_name,
		       COALESCE(u.last_name, '') as last_name
		FROM "organization_member" om
		INNER JOIN "account" a ON om.user_id = a.id
		LEFT JOIN "user" u ON u.account_id = a.id
		WHERE om.organization_id = $1
		ORDER BY om.joined_at ASC
	`

	rows, err := q.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.OrganizationMember
	for rows.Next() {
		var member models.OrganizationMember
		var userName, timezone, locale, userEmail, firstName, lastName string
		err := rows.Scan(
			&member.ID,
			&member.OrganizationID,
			&member.UserID,
			&member.Role,
			&member.JoinedAt,
			&member.UpdatedAt,
			&userName,
			&timezone,
			&locale,
			&userEmail,
			&firstName,
			&lastName,
		)
		if err != nil {
			return nil, err
		}
		// Populate extended fields
		member.Email = userEmail
		member.FirstName = firstName
		member.LastName = lastName
		member.UserName = userName
		members = append(members, member)
	}

	return members, rows.Err()
}

// UpdateOrganizationMember updates a member's role
func (q *OrganizationQueries) UpdateOrganizationMember(ctx context.Context, orgID, userID string, role models.OrganizationRole) error {
	query := `
		UPDATE "organization_member"
		SET role = $3, updated_at = $4
		WHERE organization_id = $1 AND user_id = $2
	`

	_, err := q.Exec(ctx, query, orgID, userID, role, time.Now())
	if err != nil {
		// Error returned to handler for logging with context
		return err
		
	}
	return nil
}

// RemoveOrganizationMember removes a member from an organization
func (q *OrganizationQueries) RemoveOrganizationMember(ctx context.Context, orgID, userID string) error {
	query := `
		DELETE FROM "organization_member"
		WHERE organization_id = $1 AND user_id = $2
	`

	_, err := q.Exec(ctx, query, orgID, userID)
	if err != nil {
		// Error returned to handler for logging with context
		return err
		
	}
	return nil
}

// CreateOrganizationInvitation creates a new invitation
func (q *OrganizationQueries) CreateOrganizationInvitation(ctx context.Context, invitation *models.OrganizationInvitation) error {
	query := `
		INSERT INTO "organization_invitation" (
			"id", "organization_id", "email", "role", "token", "expires_at",
			"invited_by_id", "created_at", "accepted_at"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	if invitation.CreatedAt.IsZero() {
		invitation.CreatedAt = time.Now()
	}

	_, err := q.Exec(ctx, query,
		invitation.ID,
		invitation.OrganizationID,
		invitation.Email,
		invitation.Role,
		invitation.Token,
		invitation.ExpiresAt,
		invitation.InvitedByID,
		invitation.CreatedAt,
		invitation.AcceptedAt,
	)

	if err != nil {
		// Error returned to handler for logging with context
		return err
		
	}
	return nil
}

// GetOrganizationInvitation retrieves an invitation by token
func (q *OrganizationQueries) GetOrganizationInvitation(ctx context.Context, token string) (*models.OrganizationInvitation, error) {
	query := `
		SELECT id, organization_id, email, role, token, expires_at,
		       invited_by_id, created_at, accepted_at
		FROM "organization_invitation"
		WHERE token = $1 AND expires_at > $2 AND accepted_at IS NULL
	`

	var invitation models.OrganizationInvitation
	err := q.QueryRow(ctx, query, token, time.Now()).Scan(
		&invitation.ID,
		&invitation.OrganizationID,
		&invitation.Email,
		&invitation.Role,
		&invitation.Token,
		&invitation.ExpiresAt,
		&invitation.InvitedByID,
		&invitation.CreatedAt,
		&invitation.AcceptedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil // Invitation not found or expired
		}
		// Error returned to handler for logging with context
		return nil, err
		
	}

	return &invitation, nil
}

// AcceptOrganizationInvitation marks an invitation as accepted
func (q *OrganizationQueries) AcceptOrganizationInvitation(ctx context.Context, token string, userID string) error {
	// First get the invitation
	invitation, err := q.GetOrganizationInvitation(ctx, token)
	if err != nil {
		return err
	}
	if invitation == nil {
		return fmt.Errorf("invitation not found or expired")
	}

	// Mark invitation as accepted
	query := `
		UPDATE "organization_invitation"
		SET accepted_at = $2
		WHERE token = $1
	`

	_, err = q.Exec(ctx, query, token, time.Now())
	if err != nil {
		// Error returned to handler for logging with context
		return err
		
	}

	// Add user as member
	member := &models.OrganizationMember{
		ID:             generateID(),
		OrganizationID: invitation.OrganizationID,
		UserID:         userID,
		Role:           invitation.Role,
	}

	return q.AddOrganizationMember(ctx, member)
}

// GetOrganizationInvitations retrieves all pending invitations for an organization
func (q *OrganizationQueries) GetOrganizationInvitations(ctx context.Context, orgID string) ([]models.OrganizationInvitation, error) {
	query := `
		SELECT id, organization_id, email, role, token, expires_at,
		       invited_by_id, created_at, accepted_at
		FROM "organization_invitation"
		WHERE organization_id = $1 AND accepted_at IS NULL AND expires_at > $2
		ORDER BY created_at DESC
	`

	rows, err := q.Query(ctx, query, orgID, time.Now())
	if err != nil {
		// Error returned to handler for logging with context
		return nil, err
		
	}
	defer rows.Close()

	var invitations []models.OrganizationInvitation
	for rows.Next() {
		var invitation models.OrganizationInvitation
		err := rows.Scan(
			&invitation.ID,
			&invitation.OrganizationID,
			&invitation.Email,
			&invitation.Role,
			&invitation.Token,
			&invitation.ExpiresAt,
			&invitation.InvitedByID,
			&invitation.CreatedAt,
			&invitation.AcceptedAt,
		)
		if err != nil {
			// Error returned to handler for logging with context
			return nil, err
		}
		invitations = append(invitations, invitation)
	}

	if err := rows.Err(); err != nil {
		// Error returned to handler for logging with context
		return nil, err
		
	}

	return invitations, nil
}

// DeleteOrganizationInvitation deletes an invitation
func (q *OrganizationQueries) DeleteOrganizationInvitation(ctx context.Context, orgID, email string) error {
	query := `
		DELETE FROM "organization_invitation"
		WHERE organization_id = $1 AND email = $2 AND accepted_at IS NULL
	`

	_, err := q.Exec(ctx, query, orgID, email)
	if err != nil {
		// Error returned to handler for logging with context
		return err
		
	}
	return nil
}

// Helper function to generate ID using UUID
func generateID() string {
	return uuid.New().String()
}
