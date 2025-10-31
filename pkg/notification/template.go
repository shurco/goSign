package notification

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

// TemplateEngine manages notification templates
type TemplateEngine struct {
	templates map[string]*template.Template
}

// NewTemplateEngine creates new template engine
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		templates: make(map[string]*template.Template),
	}
}

// RegisterTemplate registers a template
func (e *TemplateEngine) RegisterTemplate(name, content string) error {
	tmpl, err := template.New(name).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}
	e.templates[name] = tmpl
	return nil
}

// Render renders template with data
func (e *TemplateEngine) Render(name string, data map[string]any) (string, error) {
	tmpl, ok := e.templates[name]
	if !ok {
		// If template not found, use simple variable replacement
		return e.simpleRender(name, data), nil
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	return buf.String(), nil
}

// simpleRender performs simple variable {{variable}} replacement in text
func (e *TemplateEngine) simpleRender(content string, data map[string]any) string {
	result := content
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprint(value))
	}
	return result
}

// DefaultTemplates returns default templates
func DefaultTemplates() map[string]string {
	return map[string]string{
		"invitation": `
Hello, {{submitter_name}}!

You have been sent a document "{{document_name}}" for signing.

Click the link to sign:
{{signing_url}}

The document will be available until {{expiration_date}}.

Best regards,
{{company_name}}
`,
		"completed": `
Hello, {{submitter_name}}!

You have successfully signed the document "{{document_name}}".

The signed document is available at the link:
{{document_url}}

Best regards,
{{company_name}}
`,
		"reminder": `
Hello, {{submitter_name}}!

We remind you that the document "{{document_name}}" is waiting for your signature.

Click the link to sign:
{{signing_url}}

The document will be available until {{expiration_date}}.

Best regards,
{{company_name}}
`,
		"declined": `
Hello!

User {{submitter_name}} declined to sign the document "{{document_name}}".

Reason: {{decline_reason}}

Best regards,
{{company_name}}
`,
		"email_verification": `
Hello!

Thank you for registering with goSign.

Please verify your email address by clicking the link below:
{{verification_url}}

This link will expire in 24 hours.

If you didn't register for a goSign account, please ignore this email.

Best regards,
goSign Team
`,
		"password_reset": `
Hello!

You have requested to reset your password for your goSign account.

Click the link below to reset your password:
{{reset_url}}

This link will expire in 1 hour.

If you didn't request a password reset, please ignore this email.

Best regards,
goSign Team
`,
		"2fa_enabled": `
Hello!

Two-factor authentication (2FA) has been enabled for your goSign account.

Your account is now more secure with an additional layer of protection.

If you didn't enable 2FA, please contact support immediately.

Best regards,
goSign Team
`,
		"2fa_disabled": `
Hello!

Two-factor authentication (2FA) has been disabled for your goSign account.

If you didn't disable 2FA, please contact support immediately and change your password.

Best regards,
goSign Team
`,
		"welcome": `
Hello {{user_name}}!

Welcome to goSign! We're excited to have you on board.

You can now start creating and signing documents with ease.

Get started by exploring our features:
- Create document templates
- Send documents for signing
- Track signature progress

If you have any questions, feel free to reach out to our support team.

Best regards,
goSign Team
`,
	}
}

