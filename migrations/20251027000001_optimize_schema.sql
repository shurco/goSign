-- +goose Up
-- +goose StatementBegin

-- Create notification table (universal for email/SMS/reminders)
CREATE TABLE "public"."notification" (
  "id" uuid DEFAULT gen_random_uuid (),
  "type" varchar(20) NOT NULL, -- email, sms, reminder
  "recipient" varchar(255) NOT NULL,
  "template_name" varchar(100),
  "subject" varchar(500),
  "body" text,
  "context_json" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "status" varchar(20) NOT NULL DEFAULT 'pending', -- pending, sent, failed, cancelled
  "scheduled_at" timestamp NOT NULL DEFAULT NOW(),
  "sent_at" timestamp DEFAULT NULL,
  "related_type" varchar(50), -- submission, submitter, template, etc.
  "related_id" uuid,
  "error_message" text,
  "retry_count" int DEFAULT 0,
  "created_at" timestamp DEFAULT NOW(),
  "updated_at" timestamp DEFAULT NOW(),
  PRIMARY KEY ("id")
);
CREATE INDEX "notification_on_status_scheduled" ON "public"."notification" USING BTREE ("status", "scheduled_at");
CREATE INDEX "notification_on_related" ON "public"."notification" USING BTREE ("related_type", "related_id");

-- Create webhook table
CREATE TABLE "public"."webhook" (
  "id" uuid DEFAULT gen_random_uuid (),
  "account_id" uuid NOT NULL,
  "url" varchar(1000) NOT NULL,
  "events" jsonb NOT NULL DEFAULT '[]'::jsonb, -- array of events ["submission.created", "submission.completed", etc.]
  "secret" varchar(255) NOT NULL,
  "enabled" bool NOT NULL DEFAULT true,
  "last_triggered_at" timestamp DEFAULT NULL,
  "failure_count" int DEFAULT 0,
  "created_at" timestamp DEFAULT NOW(),
  "updated_at" timestamp DEFAULT NOW(),
  FOREIGN KEY ("account_id") REFERENCES "public"."account"("id") ON DELETE CASCADE,
  PRIMARY KEY ("id")
);
CREATE INDEX "webhook_on_account_id" ON "public"."webhook" USING BTREE ("account_id");
CREATE INDEX "webhook_on_enabled" ON "public"."webhook" USING BTREE ("enabled");

-- Create event table (universal event logging)
CREATE TABLE "public"."event" (
  "id" uuid DEFAULT gen_random_uuid (),
  "type" varchar(100) NOT NULL, -- submission.created, submitter.completed, etc.
  "actor_id" uuid, -- user_id who triggered the event
  "resource_type" varchar(50) NOT NULL, -- submission, submitter, template, etc.
  "resource_id" uuid NOT NULL,
  "metadata_json" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "ip" inet,
  "created_at" timestamp DEFAULT NOW(),
  PRIMARY KEY ("id")
);
CREATE INDEX "event_on_type" ON "public"."event" USING BTREE ("type");
CREATE INDEX "event_on_resource" ON "public"."event" USING BTREE ("resource_type", "resource_id");
CREATE INDEX "event_on_created_at" ON "public"."event" USING BTREE ("created_at" DESC);

-- Create api_key table
CREATE TABLE "public"."api_key" (
  "id" uuid DEFAULT gen_random_uuid (),
  "account_id" uuid NOT NULL,
  "name" varchar(100) NOT NULL,
  "key_hash" varchar(255) NOT NULL,
  "last_used_at" timestamp DEFAULT NULL,
  "expires_at" timestamp DEFAULT NULL,
  "created_at" timestamp DEFAULT NOW(),
  "updated_at" timestamp DEFAULT NOW(),
  FOREIGN KEY ("account_id") REFERENCES "public"."account"("id") ON DELETE CASCADE,
  PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX "api_key_on_key_hash" ON "public"."api_key" USING BTREE ("key_hash");
CREATE INDEX "api_key_on_account_id" ON "public"."api_key" USING BTREE ("account_id");

-- Extend submitter table
ALTER TABLE "public"."submitter" ADD COLUMN IF NOT EXISTS "status" varchar(20) NOT NULL DEFAULT 'pending'; -- pending, opened, completed, declined
ALTER TABLE "public"."submitter" ADD COLUMN IF NOT EXISTS "declined_at" timestamp DEFAULT NULL;
ALTER TABLE "public"."submitter" ADD COLUMN IF NOT EXISTS "metadata" jsonb NOT NULL DEFAULT '{}'::jsonb;

-- Create index for submitter status
CREATE INDEX IF NOT EXISTS "submitter_on_status" ON "public"."submitter" USING BTREE ("status");

-- Extend template table
ALTER TABLE "public"."template" ADD COLUMN IF NOT EXISTS "settings" jsonb NOT NULL DEFAULT '{}'::jsonb;
-- settings will contain: embedding_enabled, webhook_enabled, expiration_days, company_logo_id, etc.

-- Extend account table
ALTER TABLE "public"."account" ADD COLUMN IF NOT EXISTS "settings" jsonb NOT NULL DEFAULT '{}'::jsonb;
-- settings will contain: SMTP config, storage config, branding, webhooks, etc.

-- Remove redundant fields from submission (get via JOIN with template)
-- These fields duplicate template data and may become stale
ALTER TABLE "public"."submission" DROP COLUMN IF EXISTS "template_fields";
ALTER TABLE "public"."submission" DROP COLUMN IF EXISTS "template_schema";
ALTER TABLE "public"."submission" DROP COLUMN IF EXISTS "template_submitters";

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore submission fields (for rollback case)
ALTER TABLE "public"."submission" ADD COLUMN IF NOT EXISTS "template_fields" jsonb NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE "public"."submission" ADD COLUMN IF NOT EXISTS "template_schema" jsonb NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE "public"."submission" ADD COLUMN IF NOT EXISTS "template_submitters" jsonb NOT NULL DEFAULT '{}'::jsonb;

-- Remove extensions from existing tables
ALTER TABLE "public"."account" DROP COLUMN IF EXISTS "settings";
ALTER TABLE "public"."template" DROP COLUMN IF EXISTS "settings";
ALTER TABLE "public"."submitter" DROP COLUMN IF EXISTS "metadata";
ALTER TABLE "public"."submitter" DROP COLUMN IF EXISTS "declined_at";
ALTER TABLE "public"."submitter" DROP COLUMN IF EXISTS "status";

-- Remove new tables
DROP TABLE IF EXISTS "api_key";
DROP TABLE IF EXISTS "event";
DROP TABLE IF EXISTS "webhook";
DROP TABLE IF EXISTS "notification";

-- +goose StatementEnd

