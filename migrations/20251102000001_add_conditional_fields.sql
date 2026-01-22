-- +goose Up
-- +goose StatementBegin
-- Add conditions column to store field conditions
ALTER TABLE template 
ADD COLUMN IF NOT EXISTS field_conditions JSONB DEFAULT '{}'::jsonb;

-- Index for faster condition lookups
CREATE INDEX IF NOT EXISTS idx_template_field_conditions ON template USING GIN(field_conditions);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_template_field_conditions;
ALTER TABLE template DROP COLUMN IF EXISTS field_conditions;
-- +goose StatementEnd
