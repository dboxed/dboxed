-- +goose Up
-- create "s3_bucket" table
CREATE TABLE "s3_bucket" (
  "id" bigserial NOT NULL,
  "workspace_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "finalizers" text NOT NULL DEFAULT '{}',
  "reconcile_status" text NOT NULL DEFAULT 'Ok',
  "reconcile_status_details" text NOT NULL DEFAULT '',
  "endpoint" text NOT NULL,
  "bucket" text NOT NULL,
  "access_key_id" text NOT NULL,
  "secret_access_key" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "s3_bucket_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- create index "s3_bucket_workspace_bucket" to table: "s3_bucket"
CREATE INDEX "s3_bucket_workspace_bucket" ON "s3_bucket" ("workspace_id", "bucket");
-- modify "volume_provider_rustic" table
ALTER TABLE "volume_provider_rustic" ADD COLUMN "s3_bucket_id" bigint NULL, ADD COLUMN "storage_prefix" text NOT NULL, ADD CONSTRAINT "volume_provider_rustic_s3_bucket_id_fkey" FOREIGN KEY ("s3_bucket_id") REFERENCES "s3_bucket" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- drop "volume_provider_storage_s3" table
DROP TABLE "volume_provider_storage_s3";

-- +goose Down
-- reverse: drop "volume_provider_storage_s3" table
CREATE TABLE "volume_provider_storage_s3" (
  "id" bigint NOT NULL,
  "endpoint" text NOT NULL,
  "bucket" text NOT NULL,
  "access_key_id" text NOT NULL,
  "secret_access_key" text NOT NULL,
  "prefix" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "volume_provider_storage_s3_id_fkey" FOREIGN KEY ("id") REFERENCES "volume_provider" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- reverse: modify "volume_provider_rustic" table
ALTER TABLE "volume_provider_rustic" DROP CONSTRAINT "volume_provider_rustic_s3_bucket_id_fkey", DROP COLUMN "storage_prefix", DROP COLUMN "s3_bucket_id";
-- reverse: create index "s3_bucket_workspace_bucket" to table: "s3_bucket"
DROP INDEX "s3_bucket_workspace_bucket";
-- reverse: create "s3_bucket" table
DROP TABLE "s3_bucket";
