-- +goose Up
-- +goose StatementBegin
-- Add locale columns for i18n support
ALTER TABLE "user" ADD COLUMN IF NOT EXISTS preferred_locale VARCHAR(10);
ALTER TABLE template ADD COLUMN IF NOT EXISTS default_locale VARCHAR(10) DEFAULT 'en';
ALTER TABLE submission ADD COLUMN IF NOT EXISTS locale VARCHAR(10);

-- Add translations JSONB column to template for field labels
ALTER TABLE template ADD COLUMN IF NOT EXISTS translations JSONB DEFAULT '{}'::jsonb;

-- Create indexes for faster locale queries
CREATE INDEX IF NOT EXISTS idx_user_preferred_locale ON "user"(preferred_locale);
CREATE INDEX IF NOT EXISTS idx_template_default_locale ON template(default_locale);
CREATE INDEX IF NOT EXISTS idx_submission_locale ON submission(locale);
CREATE INDEX IF NOT EXISTS idx_template_translations ON template USING GIN(translations);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_template_translations;
DROP INDEX IF EXISTS idx_submission_locale;
DROP INDEX IF EXISTS idx_template_default_locale;
DROP INDEX IF EXISTS idx_user_preferred_locale;

ALTER TABLE submission DROP COLUMN IF EXISTS locale;
ALTER TABLE template DROP COLUMN IF EXISTS translations;
ALTER TABLE template DROP COLUMN IF EXISTS default_locale;
ALTER TABLE "user" DROP COLUMN IF EXISTS preferred_locale;
-- +goose StatementEnd
