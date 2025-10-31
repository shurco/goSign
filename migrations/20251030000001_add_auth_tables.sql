-- +goose Up
-- +goose StatementBegin

-- Add email verification and 2FA fields to user table
ALTER TABLE "public"."user" 
  ADD COLUMN IF NOT EXISTS "email_verified" bool NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS "email_verified_at" timestamp DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS "otp_enabled" bool NOT NULL DEFAULT false;

-- Create password reset tokens table
CREATE TABLE IF NOT EXISTS "public"."password_reset_token" (
  "id" uuid DEFAULT gen_random_uuid (),
  "user_id" uuid NOT NULL,
  "token" varchar(255) NOT NULL,
  "expires_at" timestamp NOT NULL,
  "used_at" timestamp DEFAULT NULL,
  "created_at" timestamp DEFAULT NOW(),
  FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE CASCADE,
  PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX "password_reset_token_on_token" ON "public"."password_reset_token" USING BTREE ("token");
CREATE INDEX "password_reset_token_on_user_id" ON "public"."password_reset_token" USING BTREE ("user_id");
CREATE INDEX "password_reset_token_on_expires_at" ON "public"."password_reset_token" USING BTREE ("expires_at");

-- Create email verification tokens table
CREATE TABLE IF NOT EXISTS "public"."email_verification_token" (
  "id" uuid DEFAULT gen_random_uuid (),
  "user_id" uuid NOT NULL,
  "token" varchar(255) NOT NULL,
  "expires_at" timestamp NOT NULL,
  "used_at" timestamp DEFAULT NULL,
  "created_at" timestamp DEFAULT NOW(),
  FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE CASCADE,
  PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX "email_verification_token_on_token" ON "public"."email_verification_token" USING BTREE ("token");
CREATE INDEX "email_verification_token_on_user_id" ON "public"."email_verification_token" USING BTREE ("user_id");
CREATE INDEX "email_verification_token_on_expires_at" ON "public"."email_verification_token" USING BTREE ("expires_at");

-- Create OAuth accounts table for linking OAuth providers
CREATE TABLE IF NOT EXISTS "public"."oauth_account" (
  "id" uuid DEFAULT gen_random_uuid (),
  "user_id" uuid NOT NULL,
  "provider" varchar(50) NOT NULL,
  "provider_user_id" varchar(255) NOT NULL,
  "access_token" text,
  "refresh_token" text,
  "expires_at" timestamp DEFAULT NULL,
  "created_at" timestamp DEFAULT NOW(),
  "updated_at" timestamp DEFAULT NOW(),
  FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE CASCADE,
  PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX "oauth_account_on_provider_user_id" ON "public"."oauth_account" USING BTREE ("provider", "provider_user_id");
CREATE INDEX "oauth_account_on_user_id" ON "public"."oauth_account" USING BTREE ("user_id");

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS "public"."oauth_account";
DROP TABLE IF EXISTS "public"."email_verification_token";
DROP TABLE IF EXISTS "public"."password_reset_token";

ALTER TABLE "public"."user" 
  DROP COLUMN IF EXISTS "otp_enabled",
  DROP COLUMN IF EXISTS "email_verified_at",
  DROP COLUMN IF EXISTS "email_verified";

-- +goose StatementEnd

