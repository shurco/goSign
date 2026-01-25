package queries

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/pkg/logging"
)

// EmailTemplateQueries handles email template database operations
type EmailTemplateQueries struct {
	*pgxpool.Pool
}

// GetEmailTemplate retrieves an email template by name, locale, and scope (organization or account).
// If organizationID is set, tries org template then system. If accountID is set, tries account then system.
func (q *EmailTemplateQueries) GetEmailTemplate(ctx context.Context, name string, locale string, accountID *string, organizationID *string) (*models.EmailTemplate, error) {
	if locale == "" {
		locale = "en"
	}

	if organizationID != nil && *organizationID != "" {
		if template, err := q.getTemplateByOrg(ctx, name, locale, *organizationID); err == nil {
			return template, nil
		}
		return q.GetEmailTemplate(ctx, name, locale, nil, nil) // fallback to system
	}

	if accountID != nil && *accountID != "" {
		if template, err := q.getTemplateByQuery(ctx, name, locale, *accountID); err == nil {
			return template, nil
		}
	}

	template, err := q.getTemplateByQuery(ctx, name, locale, nil)
	if err != nil {
		if err == sql.ErrNoRows && locale != "en" {
			return q.GetEmailTemplate(ctx, name, "en", nil, nil)
		}
		return nil, err
	}
	return template, nil
}

// getTemplateByOrg retrieves a template by organization_id
func (q *EmailTemplateQueries) getTemplateByOrg(ctx context.Context, name, locale, organizationID string) (*models.EmailTemplate, error) {
	query := `
		SELECT id, account_id, organization_id, name, locale, subject, content, is_system, created_at, updated_at
		FROM email_template
		WHERE name = $1 AND locale = $2 AND organization_id = $3
		LIMIT 1
	`
	return q.scanTemplateRow(ctx, query, name, locale, organizationID)
}

// getTemplateByQuery retrieves a template with a specific account_id (nil for system templates)
func (q *EmailTemplateQueries) getTemplateByQuery(ctx context.Context, name, locale string, accountID interface{}) (*models.EmailTemplate, error) {
	var query string
	var args []interface{}

	if accountID == nil {
		query = `
			SELECT id, account_id, organization_id, name, locale, subject, content, is_system, created_at, updated_at
			FROM email_template
			WHERE name = $1 AND locale = $2 AND account_id IS NULL AND organization_id IS NULL
			LIMIT 1
		`
		args = []interface{}{name, locale}
	} else {
		query = `
			SELECT id, account_id, organization_id, name, locale, subject, content, is_system, created_at, updated_at
			FROM email_template
			WHERE name = $1 AND locale = $2 AND account_id = $3
			LIMIT 1
		`
		args = []interface{}{name, locale, accountID}
	}

	var template models.EmailTemplate
	var accountIDScan, orgIDScan sql.NullString
	var subjectNull sql.NullString
	err := q.QueryRow(ctx, query, args...).Scan(
		&template.ID,
		&accountIDScan,
		&orgIDScan,
		&template.Name,
		&template.Locale,
		&subjectNull,
		&template.Content,
		&template.IsSystem,
		&template.CreatedAt,
		&template.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if accountIDScan.Valid {
		template.AccountID = &accountIDScan.String
	}
	if orgIDScan.Valid {
		template.OrganizationID = &orgIDScan.String
	}
	if subjectNull.Valid {
		template.Subject = subjectNull.String
	}
	return &template, nil
}

// scanTemplateRow runs query with args and scans one row into EmailTemplate (query must return id, account_id, organization_id, name, locale, subject, content, is_system, created_at, updated_at)
func (q *EmailTemplateQueries) scanTemplateRow(ctx context.Context, query string, args ...interface{}) (*models.EmailTemplate, error) {
	var template models.EmailTemplate
	var accountIDScan, orgIDScan sql.NullString
	var subjectNull sql.NullString
	err := q.QueryRow(ctx, query, args...).Scan(
		&template.ID,
		&accountIDScan,
		&orgIDScan,
		&template.Name,
		&template.Locale,
		&subjectNull,
		&template.Content,
		&template.IsSystem,
		&template.CreatedAt,
		&template.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if accountIDScan.Valid {
		template.AccountID = &accountIDScan.String
	}
	if orgIDScan.Valid {
		template.OrganizationID = &orgIDScan.String
	}
	if subjectNull.Valid {
		template.Subject = subjectNull.String
	}
	return &template, nil
}

// GetAllEmailTemplates retrieves all email templates for the given scope (organization, account, or system).
// When organizationID is set, returns org templates + system. When accountID is set, returns account + system.
func (q *EmailTemplateQueries) GetAllEmailTemplates(ctx context.Context, accountID *string, organizationID *string, locale *string) ([]models.EmailTemplate, error) {
	var query string
	var args []interface{}

	if organizationID != nil && *organizationID != "" {
		if locale != nil {
			query = `
				SELECT id, account_id, organization_id, name, locale, subject, content, is_system, created_at, updated_at
				FROM email_template
				WHERE locale = $1 AND (organization_id = $2 OR (organization_id IS NULL AND account_id IS NULL))
				ORDER BY name, locale, organization_id NULLS LAST, account_id NULLS LAST
			`
			args = []interface{}{*locale, *organizationID}
		} else {
			query = `
				SELECT id, account_id, organization_id, name, locale, subject, content, is_system, created_at, updated_at
				FROM email_template
				WHERE organization_id = $1 OR (organization_id IS NULL AND account_id IS NULL)
				ORDER BY name, locale, organization_id NULLS LAST, account_id NULLS LAST
			`
			args = []interface{}{*organizationID}
		}
	} else if accountID != nil {
		if locale != nil {
			query = `
				SELECT id, account_id, organization_id, name, locale, subject, content, is_system, created_at, updated_at
				FROM email_template
				WHERE locale = $1 AND (account_id = $2 OR (account_id IS NULL AND organization_id IS NULL))
				ORDER BY name, locale, account_id NULLS LAST
			`
			args = []interface{}{*locale, *accountID}
		} else {
			query = `
				SELECT id, account_id, organization_id, name, locale, subject, content, is_system, created_at, updated_at
				FROM email_template
				WHERE account_id = $1 OR (account_id IS NULL AND organization_id IS NULL)
				ORDER BY name, locale, account_id NULLS LAST
			`
			args = []interface{}{*accountID}
		}
	} else {
		if locale != nil {
			query = `
				SELECT id, account_id, organization_id, name, locale, subject, content, is_system, created_at, updated_at
				FROM email_template
				WHERE account_id IS NULL AND organization_id IS NULL AND locale = $1
				ORDER BY name, locale
			`
			args = []interface{}{*locale}
		} else {
			query = `
				SELECT id, account_id, organization_id, name, locale, subject, content, is_system, created_at, updated_at
				FROM email_template
				WHERE account_id IS NULL AND organization_id IS NULL
				ORDER BY name, locale
			`
			args = []interface{}{}
		}
	}

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		logging.Log.Err(err).
			Str("query", query).
			Interface("args", args).
			Msg("Failed to query email templates")
		return nil, fmt.Errorf("failed to query email templates: %w", err)
	}
	defer rows.Close()

	var templates []models.EmailTemplate
	for rows.Next() {
		var template models.EmailTemplate
		var accountIDNull, orgIDNull sql.NullString
		var subjectNull sql.NullString
		err := rows.Scan(
			&template.ID,
			&accountIDNull,
			&orgIDNull,
			&template.Name,
			&template.Locale,
			&subjectNull,
			&template.Content,
			&template.IsSystem,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if accountIDNull.Valid {
			template.AccountID = &accountIDNull.String
		}
		if orgIDNull.Valid {
			template.OrganizationID = &orgIDNull.String
		}
		if subjectNull.Valid {
			template.Subject = subjectNull.String
		}
		templates = append(templates, template)
	}

	return templates, nil
}

// CreateEmailTemplate creates a new email template (scoped by organization or account).
func (q *EmailTemplateQueries) CreateEmailTemplate(ctx context.Context, accountID *string, organizationID *string, template *models.EmailTemplate) error {
	if template.Locale == "" {
		template.Locale = "en"
	}

	query := `
		INSERT INTO email_template (id, account_id, organization_id, name, locale, subject, content, is_system, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	template.ID = uuid.New().String()
	now := time.Now()
	if template.CreatedAt.IsZero() {
		template.CreatedAt = now
	}
	if template.UpdatedAt.IsZero() {
		template.UpdatedAt = now
	}

	var accountIDNull, orgIDNull sql.NullString
	if organizationID != nil && *organizationID != "" {
		orgIDNull = sql.NullString{String: *organizationID, Valid: true}
	} else if accountID != nil {
		accountIDNull = sql.NullString{String: *accountID, Valid: true}
	}

	_, err := q.Exec(ctx, query,
		template.ID,
		accountIDNull,
		orgIDNull,
		template.Name,
		template.Locale,
		template.Subject,
		template.Content,
		template.IsSystem,
		template.CreatedAt,
		template.UpdatedAt,
	)

	if err != nil {
		logging.Log.Err(err)
		return err
	}

	return nil
}

// UpdateEmailTemplate updates an existing email template
func (q *EmailTemplateQueries) UpdateEmailTemplate(ctx context.Context, id string, template *models.EmailTemplate) error {
	query := `
		UPDATE email_template
		SET subject = $1, content = $2, updated_at = $3
		WHERE id = $4 AND is_system = FALSE
	`

	template.UpdatedAt = time.Now()

	result, err := q.Exec(ctx, query,
		template.Subject,
		template.Content,
		template.UpdatedAt,
		id,
	)

	if err != nil {
		logging.Log.Err(err)
		return err
	}

	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteEmailTemplate deletes an email template (only non-system templates)
func (q *EmailTemplateQueries) DeleteEmailTemplate(ctx context.Context, id string) error {
	query := `
		DELETE FROM email_template
		WHERE id = $1 AND is_system = FALSE
	`

	result, err := q.Exec(ctx, query, id)
	if err != nil {
		logging.Log.Err(err)
		return err
	}

	if result.RowsAffected() == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetEmailTemplateByID retrieves an email template by ID
func (q *EmailTemplateQueries) GetEmailTemplateByID(ctx context.Context, id string) (*models.EmailTemplate, error) {
	query := `
		SELECT id, account_id, organization_id, name, locale, subject, content, is_system, created_at, updated_at
		FROM email_template
		WHERE id = $1
	`

	var template models.EmailTemplate
	var accountIDNull, orgIDNull sql.NullString
	var subjectNull sql.NullString
	err := q.QueryRow(ctx, query, id).Scan(
		&template.ID,
		&accountIDNull,
		&orgIDNull,
		&template.Name,
		&template.Locale,
		&subjectNull,
		&template.Content,
		&template.IsSystem,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if accountIDNull.Valid {
		template.AccountID = &accountIDNull.String
	}
	if orgIDNull.Valid {
		template.OrganizationID = &orgIDNull.String
	}
	if subjectNull.Valid {
		template.Subject = subjectNull.String
	}

	return &template, nil
}
