-- +goose Up
-- +goose StatementBegin

-- Add organizations and team management functionality

-- Add type field to account table
ALTER TABLE "account" ADD COLUMN IF NOT EXISTS "type" varchar(20) NOT NULL DEFAULT 'personal';

-- Create organizations table
CREATE TABLE IF NOT EXISTS "organization" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "name" varchar(255) NOT NULL,
    "description" text,
    "owner_id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_organization_owner FOREIGN KEY ("owner_id") REFERENCES "account"("id") ON DELETE CASCADE
);

-- Create organization_members table
CREATE TABLE IF NOT EXISTS "organization_member" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "organization_id" uuid NOT NULL,
    "user_id" uuid NOT NULL,
    "role" varchar(20) NOT NULL DEFAULT 'member',
    "joined_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_organization_member_org FOREIGN KEY ("organization_id") REFERENCES "organization"("id") ON DELETE CASCADE,
    CONSTRAINT fk_organization_member_user FOREIGN KEY ("user_id") REFERENCES "account"("id") ON DELETE CASCADE,
    CONSTRAINT unique_org_user UNIQUE ("organization_id", "user_id")
);

-- Create organization_invitations table
CREATE TABLE IF NOT EXISTS "organization_invitation" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "organization_id" uuid NOT NULL,
    "email" varchar(255) NOT NULL,
    "role" varchar(20) NOT NULL DEFAULT 'member',
    "token" varchar(255) NOT NULL UNIQUE,
    "expires_at" timestamptz NOT NULL,
    "invited_by_id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "accepted_at" timestamptz,
    CONSTRAINT fk_invitation_org FOREIGN KEY ("organization_id") REFERENCES "organization"("id") ON DELETE CASCADE,
    CONSTRAINT fk_invitation_inviter FOREIGN KEY ("invited_by_id") REFERENCES "account"("id") ON DELETE CASCADE,
    CONSTRAINT unique_org_email UNIQUE ("organization_id", "email")
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_organization_owner_id ON "organization"("owner_id");
CREATE INDEX IF NOT EXISTS idx_organization_member_org_id ON "organization_member"("organization_id");
CREATE INDEX IF NOT EXISTS idx_organization_member_user_id ON "organization_member"("user_id");
CREATE INDEX IF NOT EXISTS idx_organization_invitation_org_id ON "organization_invitation"("organization_id");
CREATE INDEX IF NOT EXISTS idx_organization_invitation_token ON "organization_invitation"("token");
CREATE INDEX IF NOT EXISTS idx_organization_invitation_expires_at ON "organization_invitation"("expires_at");

-- Add check constraints for roles
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'check_member_role') THEN
        ALTER TABLE "organization_member" ADD CONSTRAINT check_member_role
            CHECK ("role" IN ('owner', 'admin', 'member', 'viewer'));
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'check_invitation_role') THEN
        ALTER TABLE "organization_invitation" ADD CONSTRAINT check_invitation_role
            CHECK ("role" IN ('owner', 'admin', 'member', 'viewer'));
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'check_account_type') THEN
        ALTER TABLE "account" ADD CONSTRAINT check_account_type
            CHECK ("type" IN ('personal', 'organization'));
    END IF;
END $$;

-- Update existing accounts to be personal type
UPDATE "account" SET "type" = 'personal' WHERE "type" IS NULL OR "type" = '';

-- Create function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at columns
DROP TRIGGER IF EXISTS update_organization_updated_at ON "organization";
CREATE TRIGGER update_organization_updated_at
    BEFORE UPDATE ON "organization"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_organization_member_updated_at ON "organization_member";
CREATE TRIGGER update_organization_member_updated_at
    BEFORE UPDATE ON "organization_member"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Note: template_folder table already exists from init_table migration with account_id
-- No need to recreate it here

-- Add library fields to template table
ALTER TABLE "public"."template"
  ADD COLUMN IF NOT EXISTS "category" varchar(100),
  ADD COLUMN IF NOT EXISTS "tags" text[],
  ADD COLUMN IF NOT EXISTS "is_favorite" boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS "preview_image_id" varchar(50);

-- Create template_favorite table for user favorites
CREATE TABLE IF NOT EXISTS "public"."template_favorite" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "template_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  UNIQUE ("template_id", "user_id")
);

-- Create indexes for template favorites
CREATE INDEX IF NOT EXISTS "template_favorite_on_template_id" ON "public"."template_favorite" USING BTREE ("template_id");
CREATE INDEX IF NOT EXISTS "template_favorite_on_user_id" ON "public"."template_favorite" USING BTREE ("user_id");

-- Create index on template category for filtering
CREATE INDEX IF NOT EXISTS "template_on_category" ON "public"."template" USING BTREE ("category");

-- Create index on template tags for searching (GIN index for array)
CREATE INDEX IF NOT EXISTS "template_on_tags" ON "public"."template" USING GIN ("tags");

-- Add organization_id column to template table
ALTER TABLE "public"."template" ADD COLUMN IF NOT EXISTS "organization_id" uuid;

-- Create index for organization_id
CREATE INDEX IF NOT EXISTS idx_template_organization_id ON "public"."template"("organization_id");

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove organization_id from template
DROP INDEX IF EXISTS idx_template_organization_id;
ALTER TABLE "public"."template" DROP COLUMN IF EXISTS "organization_id";

-- Remove template library fields
DROP INDEX IF EXISTS "public"."template_on_tags";
DROP INDEX IF EXISTS "public"."template_on_category";
DROP INDEX IF EXISTS "public"."template_favorite_on_user_id";
DROP INDEX IF EXISTS "public"."template_favorite_on_template_id";
DROP TABLE IF EXISTS "public"."template_favorite";
ALTER TABLE "public"."template"
  DROP COLUMN IF EXISTS "preview_image_id",
  DROP COLUMN IF EXISTS "is_favorite",
  DROP COLUMN IF EXISTS "tags",
  DROP COLUMN IF EXISTS "category";

-- Note: template_folder table is not removed as it was created in init_table migration

-- Remove organization triggers
DROP TRIGGER IF EXISTS update_organization_member_updated_at ON "organization_member";
DROP TRIGGER IF EXISTS update_organization_updated_at ON "organization";

-- Remove organization constraints
ALTER TABLE "account" DROP CONSTRAINT IF EXISTS check_account_type;
ALTER TABLE "organization_invitation" DROP CONSTRAINT IF EXISTS check_invitation_role;
ALTER TABLE "organization_member" DROP CONSTRAINT IF EXISTS check_member_role;

-- Remove organization indexes
DROP INDEX IF EXISTS idx_organization_invitation_expires_at;
DROP INDEX IF EXISTS idx_organization_invitation_token;
DROP INDEX IF EXISTS idx_organization_invitation_org_id;
DROP INDEX IF EXISTS idx_organization_member_user_id;
DROP INDEX IF EXISTS idx_organization_member_org_id;
DROP INDEX IF EXISTS idx_organization_owner_id;

-- Drop organization tables
DROP TABLE IF EXISTS "organization_invitation";
DROP TABLE IF EXISTS "organization_member";
DROP TABLE IF EXISTS "organization";

-- Remove account type column
ALTER TABLE "account" DROP COLUMN IF EXISTS "type";

-- Remove update function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- +goose StatementEnd

