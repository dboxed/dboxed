-- +goose Up
-- create "box" table
CREATE TABLE "box" (
  "id" bigserial NOT NULL,
  "workspace_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "finalizers" text NOT NULL DEFAULT '{}',
  "reconcile_status" text NOT NULL DEFAULT 'Initializing',
  "reconcile_status_details" text NOT NULL DEFAULT '',
  "uuid" text NOT NULL DEFAULT '',
  "name" text NOT NULL,
  "network_id" bigint NULL,
  "network_type" text NULL,
  "dboxed_version" text NOT NULL,
  "box_spec" bytea NOT NULL,
  "nkey" text NOT NULL,
  "nkey_seed" text NOT NULL,
  "machine_id" bigint NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "box_nkey_key" UNIQUE ("nkey"),
  CONSTRAINT "box_uuid_key" UNIQUE ("uuid"),
  CONSTRAINT "box_workspace_id_name_key" UNIQUE ("workspace_id", "name")
);
-- create "box_netbird" table
CREATE TABLE "box_netbird" (
  "id" bigint NOT NULL,
  "setup_key_id" text NULL,
  "setup_key" text NULL,
  PRIMARY KEY ("id")
);
-- create "box_volume_attachment" table
CREATE TABLE "box_volume_attachment" (
  "box_id" bigint NOT NULL,
  "volume_id" bigint NOT NULL,
  "root_uid" bigint NOT NULL,
  "root_gid" bigint NOT NULL,
  "root_mode" text NOT NULL,
  PRIMARY KEY ("box_id", "volume_id"),
  CONSTRAINT "box_volume_attachment_volume_id_key" UNIQUE ("volume_id")
);
-- create "change_tracking" table
CREATE TABLE "change_tracking" (
  "id" bigserial NOT NULL,
  "table_name" text NOT NULL,
  "entity_id" bigint NOT NULL,
  "time" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id")
);
-- create index "idx_table_and_id" to table: "change_tracking"
CREATE INDEX "idx_table_and_id" ON "change_tracking" ("table_name", "id");
-- create "machine" table
CREATE TABLE "machine" (
  "id" bigserial NOT NULL,
  "workspace_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "finalizers" text NOT NULL DEFAULT '{}',
  "reconcile_status" text NOT NULL DEFAULT 'Initializing',
  "reconcile_status_details" text NOT NULL DEFAULT '',
  "name" text NOT NULL,
  "machine_provider_id" bigint NOT NULL,
  "machine_provider_type" text NOT NULL,
  "box_id" bigint NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "machine_workspace_id_name_key" UNIQUE ("workspace_id", "name")
);
-- create "machine_aws" table
CREATE TABLE "machine_aws" (
  "id" bigint NOT NULL,
  "instance_type" text NOT NULL,
  "subnet_id" text NOT NULL,
  "root_volume_size" bigint NOT NULL,
  PRIMARY KEY ("id")
);
-- create "machine_aws_status" table
CREATE TABLE "machine_aws_status" (
  "id" bigint NOT NULL,
  "instance_id" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "machine_aws_status_instance_id_key" UNIQUE ("instance_id")
);
-- create "machine_hetzner" table
CREATE TABLE "machine_hetzner" (
  "id" bigint NOT NULL,
  "server_type" text NOT NULL,
  "server_location" text NOT NULL,
  PRIMARY KEY ("id")
);
-- create "machine_hetzner_status" table
CREATE TABLE "machine_hetzner_status" (
  "id" bigint NOT NULL,
  "server_id" bigint NULL,
  PRIMARY KEY ("id")
);
-- create "machine_provider" table
CREATE TABLE "machine_provider" (
  "id" bigserial NOT NULL,
  "workspace_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "finalizers" text NOT NULL DEFAULT '{}',
  "reconcile_status" text NOT NULL DEFAULT 'Initializing',
  "reconcile_status_details" text NOT NULL DEFAULT '',
  "type" text NOT NULL,
  "name" text NOT NULL,
  "ssh_key_public" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "machine_provider_workspace_id_name_key" UNIQUE ("workspace_id", "name")
);
-- create "machine_provider_aws" table
CREATE TABLE "machine_provider_aws" (
  "id" bigint NOT NULL,
  "region" text NOT NULL,
  "aws_access_key_id" text NULL,
  "aws_secret_access_key" text NULL,
  "vpc_id" text NOT NULL,
  PRIMARY KEY ("id")
);
-- create "machine_provider_aws_status" table
CREATE TABLE "machine_provider_aws_status" (
  "id" bigint NOT NULL,
  "vpc_name" text NULL,
  "vpc_cidr" text NULL,
  "security_group_id" text NULL,
  PRIMARY KEY ("id")
);
-- create "machine_provider_aws_subnet" table
CREATE TABLE "machine_provider_aws_subnet" (
  "machine_provider_id" bigint NOT NULL,
  "subnet_id" text NOT NULL,
  "subnet_name" text NULL,
  "availability_zone" text NOT NULL,
  "cidr" text NOT NULL,
  PRIMARY KEY ("machine_provider_id", "subnet_id")
);
-- create "machine_provider_hetzner" table
CREATE TABLE "machine_provider_hetzner" (
  "id" bigint NOT NULL,
  "hcloud_token" text NOT NULL,
  "hetzner_network_name" text NOT NULL,
  "robot_user" text NULL,
  "robot_password" text NULL,
  PRIMARY KEY ("id")
);
-- create "machine_provider_hetzner_status" table
CREATE TABLE "machine_provider_hetzner_status" (
  "id" bigint NOT NULL,
  "hetzner_network_id" bigint NULL,
  "hetzner_network_zone" text NULL,
  "hetzner_network_cidr" text NULL,
  "cloud_subnet_cidr" text NULL,
  "robot_subnet_cidr" text NULL,
  "robot_vswitch_id" bigint NULL,
  PRIMARY KEY ("id")
);
-- create "network" table
CREATE TABLE "network" (
  "id" bigserial NOT NULL,
  "workspace_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "finalizers" text NOT NULL DEFAULT '{}',
  "reconcile_status" text NOT NULL DEFAULT 'Initializing',
  "reconcile_status_details" text NOT NULL DEFAULT '',
  "type" text NOT NULL,
  "name" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "network_workspace_id_name_key" UNIQUE ("workspace_id", "name")
);
-- create "network_netbird" table
CREATE TABLE "network_netbird" (
  "id" bigint NOT NULL,
  "netbird_version" text NOT NULL,
  "api_url" text NOT NULL,
  "api_access_token" text NOT NULL,
  PRIMARY KEY ("id")
);
-- create "token" table
CREATE TABLE "token" (
  "id" bigserial NOT NULL,
  "workspace_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "name" text NOT NULL,
  "token" text NOT NULL,
  "for_workspace" boolean NOT NULL,
  "box_id" bigint NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "token_token_key" UNIQUE ("token"),
  CONSTRAINT "token_workspace_id_name_key" UNIQUE ("workspace_id", "name")
);
-- create "user" table
CREATE TABLE "user" (
  "id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "name" text NOT NULL,
  "email" text NULL,
  "avatar" text NULL,
  PRIMARY KEY ("id")
);
-- create "volume" table
CREATE TABLE "volume" (
  "id" bigserial NOT NULL,
  "workspace_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "finalizers" text NOT NULL DEFAULT '{}',
  "volume_provider_id" bigint NOT NULL,
  "volume_provider_type" text NOT NULL,
  "uuid" text NOT NULL,
  "name" text NOT NULL,
  "lock_id" text NULL,
  "lock_time" timestamptz NULL,
  "lock_box_uuid" text NULL,
  "latest_snapshot_id" bigint NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "volume_uuid_key" UNIQUE ("uuid"),
  CONSTRAINT "volume_workspace_id_name_key" UNIQUE ("workspace_id", "name")
);
-- create "volume_provider" table
CREATE TABLE "volume_provider" (
  "id" bigserial NOT NULL,
  "workspace_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "finalizers" text NOT NULL DEFAULT '{}',
  "reconcile_status" text NOT NULL DEFAULT 'Initializing',
  "reconcile_status_details" text NOT NULL DEFAULT '',
  "type" text NOT NULL,
  "name" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "volume_provider_workspace_id_name_key" UNIQUE ("workspace_id", "name")
);
-- create "volume_provider_rustic" table
CREATE TABLE "volume_provider_rustic" (
  "id" bigint NOT NULL,
  "storage_type" text NOT NULL,
  "password" text NOT NULL,
  PRIMARY KEY ("id")
);
-- create "volume_provider_storage_s3" table
CREATE TABLE "volume_provider_storage_s3" (
  "id" bigint NOT NULL,
  "endpoint" text NOT NULL,
  "region" text NULL,
  "bucket" text NOT NULL,
  "access_key_id" text NOT NULL,
  "secret_access_key" text NOT NULL,
  "prefix" text NOT NULL,
  PRIMARY KEY ("id")
);
-- create "volume_rustic" table
CREATE TABLE "volume_rustic" (
  "id" bigint NOT NULL,
  "fs_size" bigint NOT NULL,
  "fs_type" text NOT NULL,
  PRIMARY KEY ("id")
);
-- create "volume_rustic_status" table
CREATE TABLE "volume_rustic_status" (
  "id" bigint NOT NULL,
  PRIMARY KEY ("id")
);
-- create "volume_snapshot" table
CREATE TABLE "volume_snapshot" (
  "id" bigserial NOT NULL,
  "workspace_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "finalizers" text NOT NULL DEFAULT '{}',
  "volume_provider_id" bigint NOT NULL,
  "volume_id" bigint NULL,
  "lock_id" text NOT NULL,
  PRIMARY KEY ("id")
);
-- create "volume_snapshot_rustic" table
CREATE TABLE "volume_snapshot_rustic" (
  "id" bigint NOT NULL,
  "snapshot_id" text NOT NULL,
  "snapshot_time" timestamptz NOT NULL,
  "parent_snapshot_id" text NULL,
  "hostname" text NOT NULL,
  "files_new" integer NOT NULL,
  "files_changed" integer NOT NULL,
  "files_unmodified" integer NOT NULL,
  "total_files_processed" integer NOT NULL,
  "total_bytes_processed" integer NOT NULL,
  "dirs_new" integer NOT NULL,
  "dirs_changed" integer NOT NULL,
  "dirs_unmodified" integer NOT NULL,
  "total_dirs_processed" integer NOT NULL,
  "total_dirsize_processed" integer NOT NULL,
  "data_blobs" integer NOT NULL,
  "tree_blobs" integer NOT NULL,
  "data_added" integer NOT NULL,
  "data_added_packed" integer NOT NULL,
  "data_added_files" integer NOT NULL,
  "data_added_files_packed" integer NOT NULL,
  "data_added_trees" integer NOT NULL,
  "data_added_trees_packed" integer NOT NULL,
  "backup_start" timestamptz NOT NULL,
  "backup_end" timestamptz NOT NULL,
  "backup_duration" real NOT NULL,
  "total_duration" real NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "volume_snapshot_rustic_snapshot_id_key" UNIQUE ("snapshot_id")
);
-- create "workspace" table
CREATE TABLE "workspace" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "finalizers" text NOT NULL DEFAULT '{}',
  "reconcile_status" text NOT NULL DEFAULT 'Initializing',
  "reconcile_status_details" text NOT NULL DEFAULT '',
  "name" text NOT NULL,
  "nkey" text NOT NULL,
  "nkey_seed" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "workspace_nkey_key" UNIQUE ("nkey")
);
-- create "workspace_access" table
CREATE TABLE "workspace_access" (
  "workspace_id" bigint NOT NULL,
  "user_id" text NOT NULL,
  PRIMARY KEY ("workspace_id", "user_id")
);
-- modify "box" table
ALTER TABLE "box" ADD CONSTRAINT "box_machine_id_fkey" FOREIGN KEY ("machine_id") REFERENCES "machine" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "box_network_id_fkey" FOREIGN KEY ("network_id") REFERENCES "network" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT, ADD CONSTRAINT "box_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- modify "box_netbird" table
ALTER TABLE "box_netbird" ADD CONSTRAINT "box_netbird_id_fkey" FOREIGN KEY ("id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "box_volume_attachment" table
ALTER TABLE "box_volume_attachment" ADD CONSTRAINT "box_volume_attachment_box_id_fkey" FOREIGN KEY ("box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "box_volume_attachment_volume_id_fkey" FOREIGN KEY ("volume_id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- modify "machine" table
ALTER TABLE "machine" ADD CONSTRAINT "machine_box_id_fkey" FOREIGN KEY ("box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT, ADD CONSTRAINT "machine_machine_provider_id_fkey" FOREIGN KEY ("machine_provider_id") REFERENCES "machine_provider" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT, ADD CONSTRAINT "machine_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- modify "machine_aws" table
ALTER TABLE "machine_aws" ADD CONSTRAINT "machine_aws_id_fkey" FOREIGN KEY ("id") REFERENCES "machine" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "machine_aws_status" table
ALTER TABLE "machine_aws_status" ADD CONSTRAINT "machine_aws_status_id_fkey" FOREIGN KEY ("id") REFERENCES "machine" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "machine_hetzner" table
ALTER TABLE "machine_hetzner" ADD CONSTRAINT "machine_hetzner_id_fkey" FOREIGN KEY ("id") REFERENCES "machine" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "machine_hetzner_status" table
ALTER TABLE "machine_hetzner_status" ADD CONSTRAINT "machine_hetzner_status_id_fkey" FOREIGN KEY ("id") REFERENCES "machine" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "machine_provider" table
ALTER TABLE "machine_provider" ADD CONSTRAINT "machine_provider_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- modify "machine_provider_aws" table
ALTER TABLE "machine_provider_aws" ADD CONSTRAINT "machine_provider_aws_id_fkey" FOREIGN KEY ("id") REFERENCES "machine_provider" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "machine_provider_aws_status" table
ALTER TABLE "machine_provider_aws_status" ADD CONSTRAINT "machine_provider_aws_status_id_fkey" FOREIGN KEY ("id") REFERENCES "machine_provider" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "machine_provider_aws_subnet" table
ALTER TABLE "machine_provider_aws_subnet" ADD CONSTRAINT "machine_provider_aws_subnet_machine_provider_id_fkey" FOREIGN KEY ("machine_provider_id") REFERENCES "machine_provider" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "machine_provider_hetzner" table
ALTER TABLE "machine_provider_hetzner" ADD CONSTRAINT "machine_provider_hetzner_id_fkey" FOREIGN KEY ("id") REFERENCES "machine_provider" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "machine_provider_hetzner_status" table
ALTER TABLE "machine_provider_hetzner_status" ADD CONSTRAINT "machine_provider_hetzner_status_id_fkey" FOREIGN KEY ("id") REFERENCES "machine_provider" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "network" table
ALTER TABLE "network" ADD CONSTRAINT "network_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- modify "network_netbird" table
ALTER TABLE "network_netbird" ADD CONSTRAINT "network_netbird_id_fkey" FOREIGN KEY ("id") REFERENCES "network" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "token" table
ALTER TABLE "token" ADD CONSTRAINT "token_box_id_fkey" FOREIGN KEY ("box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "token_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "volume" table
ALTER TABLE "volume" ADD CONSTRAINT "volume_latest_snapshot_id_fkey" FOREIGN KEY ("latest_snapshot_id") REFERENCES "volume_snapshot" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT, ADD CONSTRAINT "volume_volume_provider_id_fkey" FOREIGN KEY ("volume_provider_id") REFERENCES "volume_provider" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT, ADD CONSTRAINT "volume_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- modify "volume_provider" table
ALTER TABLE "volume_provider" ADD CONSTRAINT "volume_provider_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- modify "volume_provider_rustic" table
ALTER TABLE "volume_provider_rustic" ADD CONSTRAINT "volume_provider_rustic_id_fkey" FOREIGN KEY ("id") REFERENCES "volume_provider" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "volume_provider_storage_s3" table
ALTER TABLE "volume_provider_storage_s3" ADD CONSTRAINT "volume_provider_storage_s3_id_fkey" FOREIGN KEY ("id") REFERENCES "volume_provider" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "volume_rustic" table
ALTER TABLE "volume_rustic" ADD CONSTRAINT "volume_rustic_id_fkey" FOREIGN KEY ("id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "volume_rustic_status" table
ALTER TABLE "volume_rustic_status" ADD CONSTRAINT "volume_rustic_status_id_fkey" FOREIGN KEY ("id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "volume_snapshot" table
ALTER TABLE "volume_snapshot" ADD CONSTRAINT "volume_snapshot_volume_id_fkey" FOREIGN KEY ("volume_id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "volume_snapshot_volume_provider_id_fkey" FOREIGN KEY ("volume_provider_id") REFERENCES "volume_provider" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT, ADD CONSTRAINT "volume_snapshot_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- modify "volume_snapshot_rustic" table
ALTER TABLE "volume_snapshot_rustic" ADD CONSTRAINT "volume_snapshot_rustic_id_fkey" FOREIGN KEY ("id") REFERENCES "volume_snapshot" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "workspace_access" table
ALTER TABLE "workspace_access" ADD CONSTRAINT "workspace_access_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "user" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT, ADD CONSTRAINT "workspace_access_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
-- reverse: modify "workspace_access" table
ALTER TABLE "workspace_access" DROP CONSTRAINT "workspace_access_workspace_id_fkey", DROP CONSTRAINT "workspace_access_user_id_fkey";
-- reverse: modify "volume_snapshot_rustic" table
ALTER TABLE "volume_snapshot_rustic" DROP CONSTRAINT "volume_snapshot_rustic_id_fkey";
-- reverse: modify "volume_snapshot" table
ALTER TABLE "volume_snapshot" DROP CONSTRAINT "volume_snapshot_workspace_id_fkey", DROP CONSTRAINT "volume_snapshot_volume_provider_id_fkey", DROP CONSTRAINT "volume_snapshot_volume_id_fkey";
-- reverse: modify "volume_rustic_status" table
ALTER TABLE "volume_rustic_status" DROP CONSTRAINT "volume_rustic_status_id_fkey";
-- reverse: modify "volume_rustic" table
ALTER TABLE "volume_rustic" DROP CONSTRAINT "volume_rustic_id_fkey";
-- reverse: modify "volume_provider_storage_s3" table
ALTER TABLE "volume_provider_storage_s3" DROP CONSTRAINT "volume_provider_storage_s3_id_fkey";
-- reverse: modify "volume_provider_rustic" table
ALTER TABLE "volume_provider_rustic" DROP CONSTRAINT "volume_provider_rustic_id_fkey";
-- reverse: modify "volume_provider" table
ALTER TABLE "volume_provider" DROP CONSTRAINT "volume_provider_workspace_id_fkey";
-- reverse: modify "volume" table
ALTER TABLE "volume" DROP CONSTRAINT "volume_workspace_id_fkey", DROP CONSTRAINT "volume_volume_provider_id_fkey", DROP CONSTRAINT "volume_latest_snapshot_id_fkey";
-- reverse: modify "token" table
ALTER TABLE "token" DROP CONSTRAINT "token_workspace_id_fkey", DROP CONSTRAINT "token_box_id_fkey";
-- reverse: modify "network_netbird" table
ALTER TABLE "network_netbird" DROP CONSTRAINT "network_netbird_id_fkey";
-- reverse: modify "network" table
ALTER TABLE "network" DROP CONSTRAINT "network_workspace_id_fkey";
-- reverse: modify "machine_provider_hetzner_status" table
ALTER TABLE "machine_provider_hetzner_status" DROP CONSTRAINT "machine_provider_hetzner_status_id_fkey";
-- reverse: modify "machine_provider_hetzner" table
ALTER TABLE "machine_provider_hetzner" DROP CONSTRAINT "machine_provider_hetzner_id_fkey";
-- reverse: modify "machine_provider_aws_subnet" table
ALTER TABLE "machine_provider_aws_subnet" DROP CONSTRAINT "machine_provider_aws_subnet_machine_provider_id_fkey";
-- reverse: modify "machine_provider_aws_status" table
ALTER TABLE "machine_provider_aws_status" DROP CONSTRAINT "machine_provider_aws_status_id_fkey";
-- reverse: modify "machine_provider_aws" table
ALTER TABLE "machine_provider_aws" DROP CONSTRAINT "machine_provider_aws_id_fkey";
-- reverse: modify "machine_provider" table
ALTER TABLE "machine_provider" DROP CONSTRAINT "machine_provider_workspace_id_fkey";
-- reverse: modify "machine_hetzner_status" table
ALTER TABLE "machine_hetzner_status" DROP CONSTRAINT "machine_hetzner_status_id_fkey";
-- reverse: modify "machine_hetzner" table
ALTER TABLE "machine_hetzner" DROP CONSTRAINT "machine_hetzner_id_fkey";
-- reverse: modify "machine_aws_status" table
ALTER TABLE "machine_aws_status" DROP CONSTRAINT "machine_aws_status_id_fkey";
-- reverse: modify "machine_aws" table
ALTER TABLE "machine_aws" DROP CONSTRAINT "machine_aws_id_fkey";
-- reverse: modify "machine" table
ALTER TABLE "machine" DROP CONSTRAINT "machine_workspace_id_fkey", DROP CONSTRAINT "machine_machine_provider_id_fkey", DROP CONSTRAINT "machine_box_id_fkey";
-- reverse: modify "box_volume_attachment" table
ALTER TABLE "box_volume_attachment" DROP CONSTRAINT "box_volume_attachment_volume_id_fkey", DROP CONSTRAINT "box_volume_attachment_box_id_fkey";
-- reverse: modify "box_netbird" table
ALTER TABLE "box_netbird" DROP CONSTRAINT "box_netbird_id_fkey";
-- reverse: modify "box" table
ALTER TABLE "box" DROP CONSTRAINT "box_workspace_id_fkey", DROP CONSTRAINT "box_network_id_fkey", DROP CONSTRAINT "box_machine_id_fkey";
-- reverse: create "workspace_access" table
DROP TABLE "workspace_access";
-- reverse: create "workspace" table
DROP TABLE "workspace";
-- reverse: create "volume_snapshot_rustic" table
DROP TABLE "volume_snapshot_rustic";
-- reverse: create "volume_snapshot" table
DROP TABLE "volume_snapshot";
-- reverse: create "volume_rustic_status" table
DROP TABLE "volume_rustic_status";
-- reverse: create "volume_rustic" table
DROP TABLE "volume_rustic";
-- reverse: create "volume_provider_storage_s3" table
DROP TABLE "volume_provider_storage_s3";
-- reverse: create "volume_provider_rustic" table
DROP TABLE "volume_provider_rustic";
-- reverse: create "volume_provider" table
DROP TABLE "volume_provider";
-- reverse: create "volume" table
DROP TABLE "volume";
-- reverse: create "user" table
DROP TABLE "user";
-- reverse: create "token" table
DROP TABLE "token";
-- reverse: create "network_netbird" table
DROP TABLE "network_netbird";
-- reverse: create "network" table
DROP TABLE "network";
-- reverse: create "machine_provider_hetzner_status" table
DROP TABLE "machine_provider_hetzner_status";
-- reverse: create "machine_provider_hetzner" table
DROP TABLE "machine_provider_hetzner";
-- reverse: create "machine_provider_aws_subnet" table
DROP TABLE "machine_provider_aws_subnet";
-- reverse: create "machine_provider_aws_status" table
DROP TABLE "machine_provider_aws_status";
-- reverse: create "machine_provider_aws" table
DROP TABLE "machine_provider_aws";
-- reverse: create "machine_provider" table
DROP TABLE "machine_provider";
-- reverse: create "machine_hetzner_status" table
DROP TABLE "machine_hetzner_status";
-- reverse: create "machine_hetzner" table
DROP TABLE "machine_hetzner";
-- reverse: create "machine_aws_status" table
DROP TABLE "machine_aws_status";
-- reverse: create "machine_aws" table
DROP TABLE "machine_aws";
-- reverse: create "machine" table
DROP TABLE "machine";
-- reverse: create index "idx_table_and_id" to table: "change_tracking"
DROP INDEX "idx_table_and_id";
-- reverse: create "change_tracking" table
DROP TABLE "change_tracking";
-- reverse: create "box_volume_attachment" table
DROP TABLE "box_volume_attachment";
-- reverse: create "box_netbird" table
DROP TABLE "box_netbird";
-- reverse: create "box" table
DROP TABLE "box";
