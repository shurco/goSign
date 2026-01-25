-- +goose Up
-- +goose StatementBegin

-- Add organization_id to email_template so each organization can have its own template settings
ALTER TABLE email_template
  ADD COLUMN IF NOT EXISTS organization_id uuid REFERENCES organization(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_email_template_organization_id ON email_template(organization_id)
  WHERE organization_id IS NOT NULL;

-- Unique constraint: one (name, locale) per organization
CREATE UNIQUE INDEX IF NOT EXISTS unique_org_template_name_locale
  ON email_template (organization_id, name, locale)
  WHERE organization_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS unique_org_template_name_locale;
DROP INDEX IF EXISTS idx_email_template_organization_id;
ALTER TABLE email_template DROP COLUMN IF EXISTS organization_id;

-- +goose StatementEnd
