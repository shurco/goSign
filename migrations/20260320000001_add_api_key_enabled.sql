-- +goose Up
-- Add missing enabled column to api_key table
ALTER TABLE "public"."api_key"
  ADD COLUMN IF NOT EXISTS "enabled" bool NOT NULL DEFAULT true;

-- +goose Down
ALTER TABLE "public"."api_key" DROP COLUMN IF EXISTS "enabled";
