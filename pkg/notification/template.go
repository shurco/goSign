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
	}
}

