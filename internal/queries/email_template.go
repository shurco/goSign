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

// GetEmailTemplate retrieves an email template by name, locale, and account ID
// If accountID is nil, retrieves system template
// If template for specified locale is not found, falls back to 'en' locale
// If account-specific template is not found, falls back to system template
func (q *EmailTemplateQueries) GetEmailTemplate(ctx context.Context, name string, locale string, accountID *string) (*models.EmailTemplate, error) {
	if locale == "" {
		locale = "en"
	}

	// Try account-specific template first if accountID is provided
	if accountID != nil && *accountID != "" {
		if template, err := q.getTemplateByQuery(ctx, name, locale, *accountID); err == nil {
			return template, nil
		}
	}

	// Fall back to system template
	template, err := q.getTemplateByQuery(ctx, name, locale, nil)
	if err != nil {
		if err == sql.ErrNoRows && locale != "en" {
			return q.GetEmailTemplate(ctx, name, "en", nil)
		}
		return nil, err
	}

	return template, nil
}

// getTemplateByQuery retrieves a template with a specific account_id (nil for system templates)
func (q *EmailTemplateQueries) getTemplateByQuery(ctx context.Context, name, locale string, accountID interface{}) (*models.EmailTemplate, error) {
	var query string
	var args []interface{}

	if accountID == nil {
		query = `
			SELECT id, account_id, name, locale, subject, content, is_system, created_at, updated_at
			FROM email_template
			WHERE name = $1 AND locale = $2 AND account_id IS NULL
			LIMIT 1
		`
		args = []interface{}{name, locale}
	} else {
		query = `
			SELECT id, account_id, name, locale, subject, content, is_system, created_at, updated_at
			FROM email_template
			WHERE name = $1 AND locale = $2 AND account_id = $3
			LIMIT 1
		`
		args = []interface{}{name, locale, accountID}
	}

	var template models.EmailTemplate
	var accountIDScan sql.NullString
	var subjectNull sql.NullString

	err := q.QueryRow(ctx, query, args...).Scan(
		&template.ID,
		&accountIDScan,
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
	if subjectNull.Valid {
		template.Subject = subjectNull.String
	}

	return &template, nil
}

// GetAllEmailTemplates retrieves all email templates for an account (or system templates if accountID is nil)
// Optionally filters by locale
func (q *EmailTemplateQueries) GetAllEmailTemplates(ctx context.Context, accountID *string, locale *string) ([]models.EmailTemplate, error) {
	var query string
	var args []interface{}

	if accountID != nil {
		if locale != nil {
			// Get account-specific templates and system templates for specific locale
			query = `
				SELECT id, account_id, name, locale, subject, content, is_system, created_at, updated_at
				FROM email_template
				WHERE locale = $1 AND (account_id = $2 OR account_id IS NULL)
				ORDER BY name, locale, account_id NULLS LAST
			`
			args = []interface{}{*locale, *accountID}
		} else {
			// Get all templates for account
			query = `
				SELECT id, account_id, name, locale, subject, content, is_system, created_at, updated_at
				FROM email_template
				WHERE account_id = $1 OR account_id IS NULL
				ORDER BY name, locale, account_id NULLS LAST
			`
			args = []interface{}{*accountID}
		}
	} else {
		if locale != nil {
			// Get only system templates for specific locale
			query = `
				SELECT id, account_id, name, locale, subject, content, is_system, created_at, updated_at
				FROM email_template
				WHERE account_id IS NULL AND locale = $1
				ORDER BY name, locale
			`
			args = []interface{}{*locale}
		} else {
			// Get all system templates
			query = `
				SELECT id, account_id, name, locale, subject, content, is_system, created_at, updated_at
				FROM email_template
				WHERE account_id IS NULL
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
		var accountIDNull sql.NullString
		var subjectNull sql.NullString

		err := rows.Scan(
			&template.ID,
			&accountIDNull,
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

		if subjectNull.Valid {
			template.Subject = subjectNull.String
		}

		templates = append(templates, template)
	}

	return templates, nil
}

// CreateEmailTemplate creates a new email template
func (q *EmailTemplateQueries) CreateEmailTemplate(ctx context.Context, accountID *string, template *models.EmailTemplate) error {
	if template.Locale == "" {
		template.Locale = "en" // Default to English
	}

	query := `
		INSERT INTO email_template (id, account_id, name, locale, subject, content, is_system, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	template.ID = uuid.New().String()
	now := time.Now()
	if template.CreatedAt.IsZero() {
		template.CreatedAt = now
	}
	if template.UpdatedAt.IsZero() {
		template.UpdatedAt = now
	}

	var accountIDNull sql.NullString
	if accountID != nil {
		accountIDNull = sql.NullString{String: *accountID, Valid: true}
	}

	_, err := q.Exec(ctx, query,
		template.ID,
		accountIDNull,
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
		SELECT id, account_id, name, locale, subject, content, is_system, created_at, updated_at
		FROM email_template
		WHERE id = $1
	`

	var template models.EmailTemplate
	var accountIDNull sql.NullString
	var subjectNull sql.NullString
	err := q.QueryRow(ctx, query, id).Scan(
		&template.ID,
		&accountIDNull,
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

	if subjectNull.Valid {
		template.Subject = subjectNull.String
	}

	return &template, nil
}
