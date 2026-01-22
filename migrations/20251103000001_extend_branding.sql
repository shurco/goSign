-- +goose Up
-- +goose StatementBegin
-- Extend account settings for branding (settings_jsonb already exists in account table)
-- Create branding assets table
CREATE TABLE IF NOT EXISTS branding_asset (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_id UUID NOT NULL REFERENCES account(id) ON DELETE CASCADE,
  type VARCHAR(50) NOT NULL, -- 'logo', 'favicon', 'email_header', 'watermark'
  file_path VARCHAR(255) NOT NULL,
  mime_type VARCHAR(100) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_branding_asset_account ON branding_asset(account_id);

-- Custom domains table
CREATE TABLE IF NOT EXISTS custom_domain (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_id UUID NOT NULL REFERENCES account(id) ON DELETE CASCADE,
  domain VARCHAR(255) NOT NULL UNIQUE,
  verified BOOLEAN DEFAULT FALSE,
  verification_token VARCHAR(255),
  ssl_enabled BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT NOW(),
  verified_at TIMESTAMP,
  CONSTRAINT valid_domain CHECK (domain ~* '^[a-z0-9][a-z0-9-]*[a-z0-9]\.[a-z]{2,}$')
);

CREATE INDEX IF NOT EXISTS idx_custom_domain_account ON custom_domain(account_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_custom_domain_domain ON custom_domain(domain);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_custom_domain_domain;
DROP INDEX IF EXISTS idx_custom_domain_account;
DROP TABLE IF EXISTS custom_domain;
DROP INDEX IF EXISTS idx_branding_asset_account;
DROP TABLE IF EXISTS branding_asset;
-- +goose StatementEnd
