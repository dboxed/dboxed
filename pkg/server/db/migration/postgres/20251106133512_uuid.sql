-- +goose Up
-- modify "box" table
ALTER TABLE "box" ALTER COLUMN "id" DROP DEFAULT, ALTER COLUMN "id" TYPE text, ALTER COLUMN "workspace_id" TYPE text, DROP COLUMN "uuid", ALTER COLUMN "network_id" TYPE text, ALTER COLUMN "machine_id" TYPE text;
-- drop sequence used by serial column "id"
DROP SEQUENCE IF EXISTS "box_id_seq";
-- modify "box_compose_project" table
ALTER TABLE "box_compose_project" ALTER COLUMN "box_id" TYPE text;
-- modify "box_netbird" table
ALTER TABLE "box_netbird" ALTER COLUMN "id" TYPE text;
-- modify "box_sandbox_status" table
ALTER TABLE "box_sandbox_status" ALTER COLUMN "id" TYPE text;
-- modify "box_volume_attachment" table
ALTER TABLE "box_volume_attachment" ALTER COLUMN "box_id" TYPE text, ALTER COLUMN "volume_id" TYPE text;
-- modify "change_tracking" table
ALTER TABLE "change_tracking" ALTER COLUMN "entity_id" TYPE text;
-- modify "log_line" table
ALTER TABLE "log_line" ALTER COLUMN "workspace_id" TYPE text, ALTER COLUMN "log_id" TYPE text;
-- modify "log_metadata" table
ALTER TABLE "log_metadata" ALTER COLUMN "id" DROP DEFAULT, ALTER COLUMN "id" TYPE text, ALTER COLUMN "workspace_id" TYPE text, ALTER COLUMN "box_id" TYPE text;
-- drop sequence used by serial column "id"
DROP SEQUENCE IF EXISTS "log_metadata_id_seq";
-- modify "machine" table
ALTER TABLE "machine" ALTER COLUMN "id" DROP DEFAULT, ALTER COLUMN "id" TYPE text, ALTER COLUMN "workspace_id" TYPE text, ALTER COLUMN "machine_provider_id" TYPE text, ALTER COLUMN "box_id" TYPE text;
-- drop sequence used by serial column "id"
DROP SEQUENCE IF EXISTS "machine_id_seq";
-- modify "machine_aws" table
ALTER TABLE "machine_aws" ALTER COLUMN "id" TYPE text;
-- modify "machine_aws_status" table
ALTER TABLE "machine_aws_status" ALTER COLUMN "id" TYPE text;
-- modify "machine_hetzner" table
ALTER TABLE "machine_hetzner" ALTER COLUMN "id" TYPE text;
-- modify "machine_hetzner_status" table
ALTER TABLE "machine_hetzner_status" ALTER COLUMN "id" TYPE text;
-- modify "machine_provider" table
ALTER TABLE "machine_provider" ALTER COLUMN "id" DROP DEFAULT, ALTER COLUMN "id" TYPE text, ALTER COLUMN "workspace_id" TYPE text;
-- drop sequence used by serial column "id"
DROP SEQUENCE IF EXISTS "machine_provider_id_seq";
-- modify "machine_provider_aws" table
ALTER TABLE "machine_provider_aws" ALTER COLUMN "id" TYPE text;
-- modify "machine_provider_aws_status" table
ALTER TABLE "machine_provider_aws_status" ALTER COLUMN "id" TYPE text;
-- modify "machine_provider_aws_subnet" table
ALTER TABLE "machine_provider_aws_subnet" ALTER COLUMN "machine_provider_id" TYPE text;
-- modify "machine_provider_hetzner" table
ALTER TABLE "machine_provider_hetzner" ALTER COLUMN "id" TYPE text;
-- modify "machine_provider_hetzner_status" table
ALTER TABLE "machine_provider_hetzner_status" ALTER COLUMN "id" TYPE text;
-- modify "network" table
ALTER TABLE "network" ALTER COLUMN "id" DROP DEFAULT, ALTER COLUMN "id" TYPE text, ALTER COLUMN "workspace_id" TYPE text;
-- drop sequence used by serial column "id"
DROP SEQUENCE IF EXISTS "network_id_seq";
-- modify "network_netbird" table
ALTER TABLE "network_netbird" ALTER COLUMN "id" TYPE text;
-- modify "s3_bucket" table
ALTER TABLE "s3_bucket" ALTER COLUMN "id" DROP DEFAULT, ALTER COLUMN "id" TYPE text, ALTER COLUMN "workspace_id" TYPE text;
-- drop sequence used by serial column "id"
DROP SEQUENCE IF EXISTS "s3_bucket_id_seq";
-- modify "token" table
ALTER TABLE "token" ALTER COLUMN "id" DROP DEFAULT, ALTER COLUMN "id" TYPE text, ALTER COLUMN "workspace_id" TYPE text, ALTER COLUMN "box_id" TYPE text;
-- drop sequence used by serial column "id"
DROP SEQUENCE IF EXISTS "token_id_seq";
-- modify "volume" table
ALTER TABLE "volume" ALTER COLUMN "id" DROP DEFAULT, ALTER COLUMN "id" TYPE text, ALTER COLUMN "workspace_id" TYPE text, ALTER COLUMN "volume_provider_id" TYPE text, DROP COLUMN "uuid", ALTER COLUMN "lock_box_id" TYPE text, ALTER COLUMN "latest_snapshot_id" TYPE text;
-- drop sequence used by serial column "id"
DROP SEQUENCE IF EXISTS "volume_id_seq";
-- modify "volume_provider" table
ALTER TABLE "volume_provider" ALTER COLUMN "id" DROP DEFAULT, ALTER COLUMN "id" TYPE text, ALTER COLUMN "workspace_id" TYPE text;
-- drop sequence used by serial column "id"
DROP SEQUENCE IF EXISTS "volume_provider_id_seq";
-- modify "volume_provider_rustic" table
ALTER TABLE "volume_provider_rustic" ALTER COLUMN "id" TYPE text, ALTER COLUMN "s3_bucket_id" TYPE text;
-- modify "volume_rustic" table
ALTER TABLE "volume_rustic" ALTER COLUMN "id" TYPE text;
-- modify "volume_rustic_status" table
ALTER TABLE "volume_rustic_status" ALTER COLUMN "id" TYPE text;
-- modify "volume_snapshot" table
ALTER TABLE "volume_snapshot" ALTER COLUMN "id" DROP DEFAULT, ALTER COLUMN "id" TYPE text, ALTER COLUMN "workspace_id" TYPE text, ALTER COLUMN "volume_provider_id" TYPE text, ALTER COLUMN "volume_id" TYPE text;
-- drop sequence used by serial column "id"
DROP SEQUENCE IF EXISTS "volume_snapshot_id_seq";
-- modify "volume_snapshot_rustic" table
ALTER TABLE "volume_snapshot_rustic" ALTER COLUMN "id" TYPE text;
-- modify "workspace" table
ALTER TABLE "workspace" ALTER COLUMN "id" DROP DEFAULT, ALTER COLUMN "id" TYPE text;
-- drop sequence used by serial column "id"
DROP SEQUENCE IF EXISTS "workspace_id_seq";
-- modify "workspace_access" table
ALTER TABLE "workspace_access" ALTER COLUMN "workspace_id" TYPE text;
-- modify "workspace_quotas" table
ALTER TABLE "workspace_quotas" ALTER COLUMN "workspace_id" TYPE text;
-- create "volume_mount_status" table
CREATE TABLE "volume_mount_status" (
  "id" text NOT NULL,
  "status_time" timestamptz NULL,
  "run_status" text NULL,
  "start_time" timestamptz NULL,
  "stop_time" timestamptz NULL,
  "docker_ps" bytea NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "volume_mount_status_id_fkey" FOREIGN KEY ("id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- +goose Down
-- reverse: create "volume_mount_status" table
DROP TABLE "volume_mount_status";
-- reverse: modify "workspace_quotas" table
ALTER TABLE "workspace_quotas" ALTER COLUMN "workspace_id" TYPE bigint;
-- reverse: modify "workspace_access" table
ALTER TABLE "workspace_access" ALTER COLUMN "workspace_id" TYPE bigint;
-- reverse: drop sequence used by serial column "id"
CREATE SEQUENCE IF NOT EXISTS "workspace_id_seq" OWNED BY "workspace"."id";
-- reverse: modify "workspace" table
ALTER TABLE "workspace" ALTER COLUMN "id" SET DEFAULT nextval('"workspace_id_seq"'), ALTER COLUMN "id" TYPE bigint;
-- reverse: modify "volume_snapshot_rustic" table
ALTER TABLE "volume_snapshot_rustic" ALTER COLUMN "id" TYPE bigint;
-- reverse: drop sequence used by serial column "id"
CREATE SEQUENCE IF NOT EXISTS "volume_snapshot_id_seq" OWNED BY "volume_snapshot"."id";
-- reverse: modify "volume_snapshot" table
ALTER TABLE "volume_snapshot" ALTER COLUMN "volume_id" TYPE bigint, ALTER COLUMN "volume_provider_id" TYPE bigint, ALTER COLUMN "workspace_id" TYPE bigint, ALTER COLUMN "id" SET DEFAULT nextval('"volume_snapshot_id_seq"'), ALTER COLUMN "id" TYPE bigint;
-- reverse: modify "volume_rustic_status" table
ALTER TABLE "volume_rustic_status" ALTER COLUMN "id" TYPE bigint;
-- reverse: modify "volume_rustic" table
ALTER TABLE "volume_rustic" ALTER COLUMN "id" TYPE bigint;
-- reverse: modify "volume_provider_rustic" table
ALTER TABLE "volume_provider_rustic" ALTER COLUMN "s3_bucket_id" TYPE bigint, ALTER COLUMN "id" TYPE bigint;
-- reverse: drop sequence used by serial column "id"
CREATE SEQUENCE IF NOT EXISTS "volume_provider_id_seq" OWNED BY "volume_provider"."id";
-- reverse: modify "volume_provider" table
ALTER TABLE "volume_provider" ALTER COLUMN "workspace_id" TYPE bigint, ALTER COLUMN "id" SET DEFAULT nextval('"volume_provider_id_seq"'), ALTER COLUMN "id" TYPE bigint;
-- reverse: drop sequence used by serial column "id"
CREATE SEQUENCE IF NOT EXISTS "volume_id_seq" OWNED BY "volume"."id";
-- reverse: modify "volume" table
ALTER TABLE "volume" ALTER COLUMN "latest_snapshot_id" TYPE bigint, ALTER COLUMN "lock_box_id" TYPE bigint, ADD COLUMN "uuid" text NOT NULL, ALTER COLUMN "volume_provider_id" TYPE bigint, ALTER COLUMN "workspace_id" TYPE bigint, ALTER COLUMN "id" SET DEFAULT nextval('"volume_id_seq"'), ALTER COLUMN "id" TYPE bigint;
-- reverse: drop sequence used by serial column "id"
CREATE SEQUENCE IF NOT EXISTS "token_id_seq" OWNED BY "token"."id";
-- reverse: modify "token" table
ALTER TABLE "token" ALTER COLUMN "box_id" TYPE bigint, ALTER COLUMN "workspace_id" TYPE bigint, ALTER COLUMN "id" SET DEFAULT nextval('"token_id_seq"'), ALTER COLUMN "id" TYPE bigint;
-- reverse: drop sequence used by serial column "id"
CREATE SEQUENCE IF NOT EXISTS "s3_bucket_id_seq" OWNED BY "s3_bucket"."id";
-- reverse: modify "s3_bucket" table
ALTER TABLE "s3_bucket" ALTER COLUMN "workspace_id" TYPE bigint, ALTER COLUMN "id" SET DEFAULT nextval('"s3_bucket_id_seq"'), ALTER COLUMN "id" TYPE bigint;
-- reverse: modify "network_netbird" table
ALTER TABLE "network_netbird" ALTER COLUMN "id" TYPE bigint;
-- reverse: drop sequence used by serial column "id"
CREATE SEQUENCE IF NOT EXISTS "network_id_seq" OWNED BY "network"."id";
-- reverse: modify "network" table
ALTER TABLE "network" ALTER COLUMN "workspace_id" TYPE bigint, ALTER COLUMN "id" SET DEFAULT nextval('"network_id_seq"'), ALTER COLUMN "id" TYPE bigint;
-- reverse: modify "machine_provider_hetzner_status" table
ALTER TABLE "machine_provider_hetzner_status" ALTER COLUMN "id" TYPE bigint;
-- reverse: modify "machine_provider_hetzner" table
ALTER TABLE "machine_provider_hetzner" ALTER COLUMN "id" TYPE bigint;
-- reverse: modify "machine_provider_aws_subnet" table
ALTER TABLE "machine_provider_aws_subnet" ALTER COLUMN "machine_provider_id" TYPE bigint;
-- reverse: modify "machine_provider_aws_status" table
ALTER TABLE "machine_provider_aws_status" ALTER COLUMN "id" TYPE bigint;
-- reverse: modify "machine_provider_aws" table
ALTER TABLE "machine_provider_aws" ALTER COLUMN "id" TYPE bigint;
-- reverse: drop sequence used by serial column "id"
CREATE SEQUENCE IF NOT EXISTS "machine_provider_id_seq" OWNED BY "machine_provider"."id";
-- reverse: modify "machine_provider" table
ALTER TABLE "machine_provider" ALTER COLUMN "workspace_id" TYPE bigint, ALTER COLUMN "id" SET DEFAULT nextval('"machine_provider_id_seq"'), ALTER COLUMN "id" TYPE bigint;
-- reverse: modify "machine_hetzner_status" table
ALTER TABLE "machine_hetzner_status" ALTER COLUMN "id" TYPE bigint;
-- reverse: modify "machine_hetzner" table
ALTER TABLE "machine_hetzner" ALTER COLUMN "id" TYPE bigint;
-- reverse: modify "machine_aws_status" table
ALTER TABLE "machine_aws_status" ALTER COLUMN "id" TYPE bigint;
-- reverse: modify "machine_aws" table
ALTER TABLE "machine_aws" ALTER COLUMN "id" TYPE bigint;
-- reverse: drop sequence used by serial column "id"
CREATE SEQUENCE IF NOT EXISTS "machine_id_seq" OWNED BY "machine"."id";
-- reverse: modify "machine" table
ALTER TABLE "machine" ALTER COLUMN "box_id" TYPE bigint, ALTER COLUMN "machine_provider_id" TYPE bigint, ALTER COLUMN "workspace_id" TYPE bigint, ALTER COLUMN "id" SET DEFAULT nextval('"machine_id_seq"'), ALTER COLUMN "id" TYPE bigint;
-- reverse: drop sequence used by serial column "id"
CREATE SEQUENCE IF NOT EXISTS "log_metadata_id_seq" OWNED BY "log_metadata"."id";
-- reverse: modify "log_metadata" table
ALTER TABLE "log_metadata" ALTER COLUMN "box_id" TYPE bigint, ALTER COLUMN "workspace_id" TYPE bigint, ALTER COLUMN "id" SET DEFAULT nextval('"log_metadata_id_seq"'), ALTER COLUMN "id" TYPE bigint;
-- reverse: modify "log_line" table
ALTER TABLE "log_line" ALTER COLUMN "log_id" TYPE bigint, ALTER COLUMN "workspace_id" TYPE bigint;
-- reverse: modify "change_tracking" table
ALTER TABLE "change_tracking" ALTER COLUMN "entity_id" TYPE bigint;
-- reverse: modify "box_volume_attachment" table
ALTER TABLE "box_volume_attachment" ALTER COLUMN "volume_id" TYPE bigint, ALTER COLUMN "box_id" TYPE bigint;
-- reverse: modify "box_sandbox_status" table
ALTER TABLE "box_sandbox_status" ALTER COLUMN "id" TYPE bigint;
-- reverse: modify "box_netbird" table
ALTER TABLE "box_netbird" ALTER COLUMN "id" TYPE bigint;
-- reverse: modify "box_compose_project" table
ALTER TABLE "box_compose_project" ALTER COLUMN "box_id" TYPE bigint;
-- reverse: drop sequence used by serial column "id"
CREATE SEQUENCE IF NOT EXISTS "box_id_seq" OWNED BY "box"."id";
-- reverse: modify "box" table
ALTER TABLE "box" ALTER COLUMN "machine_id" TYPE bigint, ALTER COLUMN "network_id" TYPE bigint, ADD COLUMN "uuid" text NOT NULL DEFAULT '', ALTER COLUMN "workspace_id" TYPE bigint, ALTER COLUMN "id" SET DEFAULT nextval('"box_id_seq"'), ALTER COLUMN "id" TYPE bigint;
