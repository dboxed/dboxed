-- +goose Up
-- modify "box" table
ALTER TABLE "box" ADD COLUMN "change_seq" bigint NOT NULL DEFAULT 0;
-- create index "box_change_seq" to table: "box"
CREATE INDEX "box_change_seq" ON "box" ("change_seq");
-- modify "box_netbird" table
ALTER TABLE "box_netbird" ADD COLUMN "change_seq" bigint NOT NULL DEFAULT 0;
-- create index "box_netbird_change_seq" to table: "box_netbird"
CREATE INDEX "box_netbird_change_seq" ON "box_netbird" ("change_seq");
-- modify "load_balancer" table
ALTER TABLE "load_balancer" ADD COLUMN "change_seq" bigint NOT NULL DEFAULT 0;
-- create index "load_balancer_change_seq" to table: "load_balancer"
CREATE INDEX "load_balancer_change_seq" ON "load_balancer" ("change_seq");
-- modify "machine" table
ALTER TABLE "machine" ADD COLUMN "change_seq" bigint NOT NULL DEFAULT 0;
-- create index "machine_change_seq" to table: "machine"
CREATE INDEX "machine_change_seq" ON "machine" ("change_seq");
-- modify "machine_aws" table
ALTER TABLE "machine_aws" ADD COLUMN "change_seq" bigint NOT NULL DEFAULT 0;
-- create index "machine_aws_change_seq" to table: "machine_aws"
CREATE INDEX "machine_aws_change_seq" ON "machine_aws" ("change_seq");
-- modify "machine_hetzner" table
ALTER TABLE "machine_hetzner" ADD COLUMN "change_seq" bigint NOT NULL DEFAULT 0;
-- create index "machine_hetzner_change_seq" to table: "machine_hetzner"
CREATE INDEX "machine_hetzner_change_seq" ON "machine_hetzner" ("change_seq");
-- modify "machine_provider" table
ALTER TABLE "machine_provider" ADD COLUMN "change_seq" bigint NOT NULL DEFAULT 0;
-- create index "machine_provider_change_seq" to table: "machine_provider"
CREATE INDEX "machine_provider_change_seq" ON "machine_provider" ("change_seq");
-- modify "network" table
ALTER TABLE "network" ADD COLUMN "change_seq" bigint NOT NULL DEFAULT 0;
-- create index "network_change_seq" to table: "network"
CREATE INDEX "network_change_seq" ON "network" ("change_seq");
-- modify "s3_bucket" table
ALTER TABLE "s3_bucket" ADD COLUMN "change_seq" bigint NOT NULL DEFAULT 0;
-- create index "s3_bucket_change_seq" to table: "s3_bucket"
CREATE INDEX "s3_bucket_change_seq" ON "s3_bucket" ("change_seq");
-- modify "volume" table
ALTER TABLE "volume" ADD COLUMN "change_seq" bigint NOT NULL DEFAULT 0;
-- create index "volume_change_seq" to table: "volume"
CREATE INDEX "volume_change_seq" ON "volume" ("change_seq");
-- modify "volume_provider" table
ALTER TABLE "volume_provider" ADD COLUMN "change_seq" bigint NOT NULL DEFAULT 0;
-- create index "volume_provider_change_seq" to table: "volume_provider"
CREATE INDEX "volume_provider_change_seq" ON "volume_provider" ("change_seq");
-- modify "workspace" table
ALTER TABLE "workspace" ADD COLUMN "change_seq" bigint NOT NULL DEFAULT 0;
-- create index "workspace_change_seq" to table: "workspace"
CREATE INDEX "workspace_change_seq" ON "workspace" ("change_seq");
-- drop "change_tracking" table
DROP TABLE "change_tracking";

-- +goose Down
-- reverse: drop "change_tracking" table
CREATE TABLE "change_tracking" (
  "id" bigserial NOT NULL,
  "table_name" text NOT NULL,
  "entity_id" text NOT NULL,
  "time" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id")
);
CREATE INDEX "idx_table_and_id" ON "change_tracking" ("table_name", "id");
-- reverse: create index "workspace_change_seq" to table: "workspace"
DROP INDEX "workspace_change_seq";
-- reverse: modify "workspace" table
ALTER TABLE "workspace" DROP COLUMN "change_seq";
-- reverse: create index "volume_provider_change_seq" to table: "volume_provider"
DROP INDEX "volume_provider_change_seq";
-- reverse: modify "volume_provider" table
ALTER TABLE "volume_provider" DROP COLUMN "change_seq";
-- reverse: create index "volume_change_seq" to table: "volume"
DROP INDEX "volume_change_seq";
-- reverse: modify "volume" table
ALTER TABLE "volume" DROP COLUMN "change_seq";
-- reverse: create index "s3_bucket_change_seq" to table: "s3_bucket"
DROP INDEX "s3_bucket_change_seq";
-- reverse: modify "s3_bucket" table
ALTER TABLE "s3_bucket" DROP COLUMN "change_seq";
-- reverse: create index "network_change_seq" to table: "network"
DROP INDEX "network_change_seq";
-- reverse: modify "network" table
ALTER TABLE "network" DROP COLUMN "change_seq";
-- reverse: create index "machine_provider_change_seq" to table: "machine_provider"
DROP INDEX "machine_provider_change_seq";
-- reverse: modify "machine_provider" table
ALTER TABLE "machine_provider" DROP COLUMN "change_seq";
-- reverse: create index "machine_hetzner_change_seq" to table: "machine_hetzner"
DROP INDEX "machine_hetzner_change_seq";
-- reverse: modify "machine_hetzner" table
ALTER TABLE "machine_hetzner" DROP COLUMN "change_seq";
-- reverse: create index "machine_aws_change_seq" to table: "machine_aws"
DROP INDEX "machine_aws_change_seq";
-- reverse: modify "machine_aws" table
ALTER TABLE "machine_aws" DROP COLUMN "change_seq";
-- reverse: create index "machine_change_seq" to table: "machine"
DROP INDEX "machine_change_seq";
-- reverse: modify "machine" table
ALTER TABLE "machine" DROP COLUMN "change_seq";
-- reverse: create index "load_balancer_change_seq" to table: "load_balancer"
DROP INDEX "load_balancer_change_seq";
-- reverse: modify "load_balancer" table
ALTER TABLE "load_balancer" DROP COLUMN "change_seq";
-- reverse: create index "box_netbird_change_seq" to table: "box_netbird"
DROP INDEX "box_netbird_change_seq";
-- reverse: modify "box_netbird" table
ALTER TABLE "box_netbird" DROP COLUMN "change_seq";
-- reverse: create index "box_change_seq" to table: "box"
DROP INDEX "box_change_seq";
-- reverse: modify "box" table
ALTER TABLE "box" DROP COLUMN "change_seq";
