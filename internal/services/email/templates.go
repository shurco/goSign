package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"github.com/shurco/gosign/internal/models"
	"github.com/shurco/gosign/internal/queries"
)

// EmailData contains data for email template rendering
type EmailData struct {
	Branding      models.BrandingSettings
	RecipientName string
	DocumentName  string
	SigningLink   string
	ExpiresAt     string
	SenderName    string
	CustomMessage string
}

// EmailTemplateService handles email template operations
type EmailTemplateService struct {
	emailTemplateQueries *queries.EmailTemplateQueries
}

// NewEmailTemplateService creates a new email template service
func NewEmailTemplateService(emailTemplateQueries *queries.EmailTemplateQueries) *EmailTemplateService {
	return &EmailTemplateService{
		emailTemplateQueries: emailTemplateQueries,
	}
}

// RenderTemplate renders email with branding from database templates
// locale defaults to "en" if empty
func (s *EmailTemplateService) RenderTemplate(ctx context.Context, templateName string, locale string, data EmailData, accountID *string) (string, error) {
	if locale == "" {
		locale = "en" // Default to English
	}

	// Get base template
	baseTemplate, err := s.emailTemplateQueries.GetEmailTemplate(ctx, "base", locale, accountID)
	if err != nil {
		return "", fmt.Errorf("failed to get base template: %w", err)
	}

	// Get content template
	contentTemplate, err := s.emailTemplateQueries.GetEmailTemplate(ctx, templateName, locale, accountID)
	if err != nil {
		return "", fmt.Errorf("failed to get content template %s: %w", templateName, err)
	}

	// Parse base template
	baseTmpl, err := template.New("base").Parse(baseTemplate.Content)
	if err != nil {
		return "", fmt.Errorf("failed to parse base template: %w", err)
	}

	// Parse content template and associate with base
	contentTmpl, err := template.Must(baseTmpl.Clone()).Parse(contentTemplate.Content)
	if err != nil {
		return "", fmt.Errorf("failed to parse content template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := contentTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute email template: %w", err)
	}

	return buf.String(), nil
}

// RenderTemplate is a legacy function that reads from files (for backward compatibility)
// It's kept for any code that might still use it, but new code should use EmailTemplateService
func RenderTemplate(templateName string, data EmailData) (string, error) {
	// This is a fallback that reads from files if DB is not available
	// In production, this should be removed once all code uses EmailTemplateService
	return "", fmt.Errorf("RenderTemplate from files is deprecated, use EmailTemplateService instead")
}
