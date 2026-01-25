-- +goose Up
-- +goose StatementBegin

-- Reorganize settings: global settings in setting table, organization settings in account.settings

-- 1. Extend setting table to store key-value pairs for global settings
ALTER TABLE "public"."setting" 
  DROP COLUMN IF EXISTS "name",
  ADD COLUMN IF NOT EXISTS "key" VARCHAR(255) NOT NULL UNIQUE,
  ADD COLUMN IF NOT EXISTS "value" JSONB NOT NULL DEFAULT '{}'::jsonb,
  ADD COLUMN IF NOT EXISTS "category" VARCHAR(100) NOT NULL DEFAULT 'general';

-- Create index for category lookup
CREATE INDEX IF NOT EXISTS "setting_on_category" ON "public"."setting" USING BTREE ("category");
CREATE INDEX IF NOT EXISTS "setting_on_key" ON "public"."setting" USING BTREE ("key");

-- 2. Update email_template to support global templates (account_id = NULL)
-- Remove the constraint that requires account_id, allow NULL for global templates
ALTER TABLE email_template 
  DROP CONSTRAINT IF EXISTS unique_template_name_per_account_locale;

-- Create new constraint that allows NULL account_id for global templates
-- Global templates: account_id IS NULL, unique by (name, locale)
-- Account templates: account_id IS NOT NULL, unique by (account_id, name, locale)
CREATE UNIQUE INDEX IF NOT EXISTS unique_global_template_name_locale 
  ON email_template (name, locale) 
  WHERE account_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS unique_account_template_name_locale 
  ON email_template (account_id, name, locale) 
  WHERE account_id IS NOT NULL;

-- 3. Insert default global settings with values from config
-- These values match the defaults from gosign.toml and config.Default()
INSERT INTO "public"."setting" ("key", "value", "category") VALUES
  ('smtp', '{
    "provider": "smtp",
    "smtp_host": "localhost",
    "smtp_port": "1025",
    "smtp_user": "",
    "smtp_pass": "",
    "from_email": "noreply@gosign.local",
    "from_name": "goSign"
  }'::jsonb, 'email'),
  ('sms', '{
    "twilio_enabled": "false",
    "twilio_account_sid": "",
    "twilio_auth_token": "",
    "twilio_from_number": ""
  }'::jsonb, 'sms'),
  ('storage', '{"provider": "local"}'::jsonb, 'storage'),
  -- Geolocation: base_dir and db_path are hardcoded in code (./base and ./base/GeoLite2-City.mmdb)
  -- Only configurable parameters are stored here (download_url, maxmind_license_key, download_method)
  ('geolocation', '{
    "download_url": "https://git.io/GeoLite2-City.mmdb"
  }'::jsonb, 'geolocation')
ON CONFLICT ("key") DO NOTHING;

-- 4. Ensure account.settings structure supports organization-specific settings
-- (webhooks, api_keys, branding are already in separate tables, but general settings can be in account.settings)
-- This is already handled by existing account.settings jsonb column

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove indexes
DROP INDEX IF EXISTS "setting_on_key";
DROP INDEX IF EXISTS "setting_on_category";
DROP INDEX IF EXISTS unique_account_template_name_locale;
DROP INDEX IF EXISTS unique_global_template_name_locale;

-- Restore original email_template constraint
ALTER TABLE email_template 
  ADD CONSTRAINT unique_template_name_per_account_locale 
  UNIQUE (account_id, name, locale);

-- Restore setting table structure: add name from key, then drop key/value/category
ALTER TABLE "public"."setting" ADD COLUMN IF NOT EXISTS "name" VARCHAR(255);
UPDATE "public"."setting" SET name = "key" WHERE "key" IS NOT NULL;
UPDATE "public"."setting" SET name = 'general' WHERE name IS NULL;
ALTER TABLE "public"."setting" ALTER COLUMN "name" SET NOT NULL;
ALTER TABLE "public"."setting"
  DROP COLUMN IF EXISTS "category",
  DROP COLUMN IF EXISTS "value",
  DROP COLUMN IF EXISTS "key";

-- +goose StatementEnd
