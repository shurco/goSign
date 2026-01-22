-- +goose Up
-- +goose StatementBegin
-- Create email_templates table
CREATE TABLE IF NOT EXISTS email_template (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_id UUID REFERENCES account(id) ON DELETE CASCADE,
  name VARCHAR(100) NOT NULL, -- 'base', 'invitation', 'reminder', 'completed'
  subject VARCHAR(255), -- Email subject (optional, for content templates)
  content TEXT NOT NULL, -- HTML template content
  is_system BOOLEAN DEFAULT FALSE, -- System templates cannot be deleted
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  CONSTRAINT unique_template_name_per_account UNIQUE (account_id, name)
);

CREATE INDEX IF NOT EXISTS idx_email_template_account ON email_template(account_id);
CREATE INDEX IF NOT EXISTS idx_email_template_name ON email_template(name);

-- Insert default system templates
-- Base template (system-wide, account_id is NULL)
INSERT INTO email_template (name, content, is_system) VALUES
('base', '<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.DocumentName}}</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            text-align: center;
            padding: 20px 0;
            border-bottom: 2px solid {{.Branding.PrimaryColor}};
        }
        .header img {
            max-height: 60px;
        }
        .content {
            padding: 20px 0;
        }
        .button {
            display: inline-block;
            padding: 12px 24px;
            background-color: {{.Branding.PrimaryColor}};
            color: white;
            text-decoration: none;
            border-radius: 4px;
            margin: 20px 0;
        }
        .footer {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #eee;
            font-size: 12px;
            color: #666;
            text-align: center;
        }
        {{if .Branding.CustomCSS}}{{.Branding.CustomCSS}}{{end}}
    </style>
</head>
<body>
    <div class="header">
        {{if .Branding.LogoURL}}
        <img src="{{.Branding.LogoURL}}" alt="{{.Branding.CompanyName}}">
        {{else}}
        <h1>{{.Branding.CompanyName}}</h1>
        {{end}}
    </div>
    
    <div class="content">
        {{template "content" .}}
    </div>
    
    <div class="footer">
        {{if .Branding.EmailFooterText}}{{.Branding.EmailFooterText}}{{else}}This email was sent by {{.Branding.CompanyName}}{{end}}
        {{if .Branding.TermsURL}}<br><a href="{{.Branding.TermsURL}}">Terms of Service</a>{{end}}
        {{if .Branding.PrivacyURL}}<a href="{{.Branding.PrivacyURL}}">Privacy Policy</a>{{end}}
    </div>
</body>
</html>', TRUE)
ON CONFLICT ON CONSTRAINT unique_template_name_per_account DO NOTHING;

-- Invitation template
INSERT INTO email_template (name, subject, content, is_system) VALUES
('invitation', 'You have been invited to sign a document', '{{define "content"}}
<p>Hello {{.RecipientName}},</p>

<p>You have been invited to sign a document: <strong>{{.DocumentName}}</strong></p>

{{if .CustomMessage}}
<p>{{.CustomMessage}}</p>
{{end}}

<p>Please click the button below to review and sign the document:</p>

<p style="text-align: center;">
    <a href="{{.SigningLink}}" class="button">Sign Document</a>
</p>

{{if .ExpiresAt}}
<p><small>This invitation expires on {{.ExpiresAt}}</small></p>
{{end}}

<p>If you have any questions, please contact {{.SenderName}}.</p>
{{end}}', TRUE)
ON CONFLICT ON CONSTRAINT unique_template_name_per_account DO NOTHING;

-- Reminder template
INSERT INTO email_template (name, subject, content, is_system) VALUES
('reminder', 'Reminder: Document awaiting your signature', '{{define "content"}}
<p>Hello {{.RecipientName}},</p>

<p>This is a reminder that you have a pending document to sign: <strong>{{.DocumentName}}</strong></p>

<p>Please click the button below to complete your signature:</p>

<p style="text-align: center;">
    <a href="{{.SigningLink}}" class="button">Sign Document</a>
</p>

{{if .ExpiresAt}}
<p><small>This invitation expires on {{.ExpiresAt}}</small></p>
{{end}}

<p>Thank you for your attention.</p>
{{end}}', TRUE)
ON CONFLICT ON CONSTRAINT unique_template_name_per_account DO NOTHING;

-- Completed template
INSERT INTO email_template (name, subject, content, is_system) VALUES
('completed', 'Document signed successfully', '{{define "content"}}
<p>Hello {{.RecipientName}},</p>

<p>The document <strong>{{.DocumentName}}</strong> has been completed and signed by all parties.</p>

<p>You can download the completed document from your dashboard.</p>

<p>Thank you for using our service.</p>
{{end}}', TRUE)
ON CONFLICT ON CONSTRAINT unique_template_name_per_account DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_email_template_name;
DROP INDEX IF EXISTS idx_email_template_account;
DROP TABLE IF EXISTS email_template;
-- +goose StatementEnd
