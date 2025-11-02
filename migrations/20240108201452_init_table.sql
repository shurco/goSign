-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE "public"."setting" (
  "id" uuid DEFAULT gen_random_uuid (),
  "name" varchar(255) NOT NULL,
  "created_at" timestamp DEFAULT NOW(),
  "updated_at" timestamp DEFAULT NOW(),
  PRIMARY KEY ("id")
);

CREATE TABLE "public"."account" (
  "id" uuid DEFAULT gen_random_uuid (),
  "name" varchar NOT NULL,
  "timezone" varchar NOT NULL,
  "locale" varchar NOT NULL,
  "created_at" timestamp DEFAULT NOW(),
  "updated_at" timestamp DEFAULT NOW(),
  PRIMARY KEY ("id")
);

CREATE TABLE "public"."user" (
  "id" uuid DEFAULT gen_random_uuid (),
  "first_name" varchar(255),
  "last_name" varchar(255),
  "email" varchar(255) NOT NULL,
  "role" int4 NOT NULL DEFAULT 1,
  "password" varchar(255) NOT NULL,
  "account_id" uuid NOT NULL,
  "sign_in_count" bool NOT NULL DEFAULT false,
  "current_sign_in_at" timestamp DEFAULT NULL,
  "last_sign_in_at" timestamp DEFAULT NULL,
  "current_sign_in_ip" inet DEFAULT NULL,
  "last_sign_in_ip" inet DEFAULT NULL,
  "otp_secret" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "consumed_timestep" integer DEFAULT NULL,
  "locked_at" timestamp DEFAULT NULL,
  "archived_at" timestamp DEFAULT NULL,
  "created_at" timestamp DEFAULT NOW(),
  "updated_at" timestamp DEFAULT NOW(),
  FOREIGN KEY ("account_id") REFERENCES "public"."account"("id") ON DELETE CASCADE,
  PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX "users_on_email" ON "public"."user" USING BTREE ("email");
CREATE INDEX "users_on_account_id" ON "public"."user" USING BTREE ("account_id");

CREATE TABLE "public"."template_folder" (
  "id" uuid DEFAULT gen_random_uuid (),
  "name" varchar(255) NOT NULL,
  "account_id" uuid NOT NULL,
  "archived_at" timestamp DEFAULT NULL,
  "created_at" timestamp DEFAULT NOW(),
  "updated_at" timestamp DEFAULT NOW(),
  FOREIGN KEY ("account_id") REFERENCES "public"."account"("id"),
  PRIMARY KEY ("id")
);
CREATE INDEX "template_folder_on_account_id" ON "public"."template_folder" USING BTREE ("account_id");

CREATE TABLE "public"."template" (
  "id" uuid DEFAULT gen_random_uuid (),
  "folder_id" uuid,
  "slug" varchar(255) NOT NULL,
  "name" varchar(255) NOT NULL,
  "schema" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "fields" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "submitters" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "source" varchar(255) NOT NULL,
  "application_key" varchar,
  "archived_at" timestamp DEFAULT NULL,
  "created_at" timestamp DEFAULT NOW(),
  "updated_at" timestamp DEFAULT NOW(),
  FOREIGN KEY ("folder_id") REFERENCES "public"."template_folder"("id"),
  PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX "template_on_slug" ON "public"."template" USING BTREE ("slug");
CREATE INDEX "template_on_folder_id" ON "public"."template" USING BTREE ("folder_id");

CREATE TABLE "public"."submission" (
  "id" uuid DEFAULT gen_random_uuid (),
  "template_id" uuid NOT NULL,
  "created_by_user_id" uuid,
  "slug" varchar NOT NULL,
  "template_fields" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "template_schema" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "template_submitters" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "source" varchar NOT NULL,
  "submitters_order" varchar NOT NULL,
  "preferences" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "archived_at" timestamp DEFAULT NULL,
  "created_at" timestamp DEFAULT NOW(),
  "updated_at" timestamp DEFAULT NOW(),
  FOREIGN KEY ("created_by_user_id") REFERENCES "public"."user"("id"),
  FOREIGN KEY ("template_id") REFERENCES "public"."template"("id"),
  PRIMARY KEY ("id")
);
CREATE INDEX "submission_on_template_id" ON "public"."submission" USING BTREE ("template_id");
CREATE UNIQUE INDEX "submission_on_slug" ON "public"."submission" USING BTREE ("slug");
CREATE INDEX "submission_on_created_by_user_id" ON "public"."submission" USING BTREE ("created_by_user_id");

CREATE TABLE "public"."submitter" (
  "id" uuid DEFAULT gen_random_uuid (),
  "submission_id" uuid NOT NULL,
  "name" varchar,
  "email" varchar(255),
  "phone" varchar,
  "slug" varchar(255) NOT NULL,
  "values" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "ua" varchar,
  "ip" inet,
  "application_key" varchar,
  "preferences" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "sented_at" timestamp,
  "opened_at" timestamp,
  "completed_at" timestamp,
  "created_at" timestamp DEFAULT NOW(),
  "updated_at" timestamp DEFAULT NOW(),
  FOREIGN KEY ("submission_id") REFERENCES "public"."submission"("id"),
  PRIMARY KEY ("id")
);
CREATE INDEX "submitter_on_submission_id" ON "public"."submitter" USING BTREE ("submission_id");
CREATE UNIQUE INDEX "submitter_on_slug" ON "public"."submitter" USING BTREE ("slug");
CREATE INDEX "submitter_on_email" ON "public"."submitter" USING BTREE ("email");

CREATE TABLE "public"."storage_blob" (
    "id" uuid DEFAULT gen_random_uuid (),
    "filename" varchar NOT NULL,
    "content_type" varchar,
    "metadata" jsonb NOT NULL DEFAULT '{}'::jsonb,
    "byte_size" int8 NOT NULL,
    PRIMARY KEY ("id")
);

CREATE TABLE "public"."storage_attachment" (
  "id" uuid DEFAULT gen_random_uuid (),
  "blob_id" uuid NOT NULL,
  "record_type" varchar NOT NULL,
  "record_id" uuid NOT NULL,
  "name" varchar NOT NULL,
  "service_name" varchar NOT NULL,
  "created_at" timestamp DEFAULT NOW(),
  FOREIGN KEY ("blob_id") REFERENCES "public"."storage_blob"("id"),
  PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX "storage_attachment_uniqueness" ON "public"."storage_attachment" USING BTREE ("record_type","record_id","name","blob_id");
CREATE INDEX "storage_attachment_on_blob_id" ON "public"."storage_attachment" USING BTREE ("blob_id");

CREATE TABLE "public"."country" (
  "code" varchar(4) NOT NULL,
  "name" varchar(255) NOT NULL,
  PRIMARY KEY ("code")
);

CREATE TABLE "public"."trust_list" (
  "list" varchar(10) NOT NULL,
  "name" varchar NOT NULL,
  "aki" varchar NOT NULL,
  "ski" varchar NOT NULL,
  "created_at" timestamp DEFAULT NOW()
);
CREATE INDEX "trust_list_aki" ON "public"."trust_list" USING BTREE ("aki");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "trust_list";
DROP TABLE IF EXISTS "country";
DROP TABLE IF EXISTS "storage_attachment";
DROP TABLE IF EXISTS "storage_blob";
DROP TABLE IF EXISTS "submitter";
DROP TABLE IF EXISTS "submission";
DROP TABLE IF EXISTS "template";
DROP TABLE IF EXISTS "template_folder";
DROP TABLE IF EXISTS "user";
DROP TABLE IF EXISTS "account";
DROP TABLE IF EXISTS "setting";
DROP EXTENSION IF EXISTS "pgcrypto";
-- +goose StatementEnd
