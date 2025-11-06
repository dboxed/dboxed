-- +goose Up
-- create "box" table
CREATE TABLE "box" (
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "finalizers" text NOT NULL DEFAULT '{}',
  "reconcile_status" text NOT NULL DEFAULT 'Initializing',
  "reconcile_status_details" text NOT NULL DEFAULT '',
  "name" text NOT NULL,
  "network_id" text NULL,
  "network_type" text NULL,
  "dboxed_version" text NOT NULL,
  "machine_id" text NULL,
  "desired_state" text NOT NULL DEFAULT 'up',
  PRIMARY KEY ("id"),
  CONSTRAINT "box_workspace_id_name_key" UNIQUE ("workspace_id", "name")
);
-- create "box_compose_project" table
CREATE TABLE "box_compose_project" (
  "box_id" text NOT NULL,
  "name" text NOT NULL,
  "compose_project" text NOT NULL,
  PRIMARY KEY ("box_id", "name")
);
-- create "box_netbird" table
CREATE TABLE "box_netbird" (
  "id" text NOT NULL,
  "reconcile_status" text NOT NULL DEFAULT 'Initializing',
  "reconcile_status_details" text NOT NULL DEFAULT '',
  "setup_key_id" text NULL,
  "setup_key" text NULL,
  PRIMARY KEY ("id")
);
-- create "box_sandbox_status" table
CREATE TABLE "box_sandbox_status" (
  "id" text NOT NULL,
  "status_time" timestamptz NULL,
  "run_status" text NULL,
  "start_time" timestamptz NULL,
  "stop_time" timestamptz NULL,
  "docker_ps" bytea NULL,
  PRIMARY KEY ("id")
);
-- create "box_volume_attachment" table
CREATE TABLE "box_volume_attachment" (
  "box_id" text NOT NULL,
  "volume_id" text NOT NULL,
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
  "entity_id" text NOT NULL,
  "time" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id")
);
-- create index "idx_table_and_id" to table: "change_tracking"
CREATE INDEX "idx_table_and_id" ON "change_tracking" ("table_name", "id");
-- create "log_line" table
CREATE TABLE "log_line" (
  "id" bigserial NOT NULL,
  "workspace_id" text NOT NULL,
  "log_id" text NOT NULL,
  "time" timestamptz NOT NULL,
  "line" text NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "log_line_log_id_and_id" to table: "log_line"
CREATE INDEX "log_line_log_id_and_id" ON "log_line" ("log_id", "id");
-- create index "log_line_time_index" to table: "log_line"
CREATE INDEX "log_line_time_index" ON "log_line" ("log_id", "time");
-- create "log_metadata" table
CREATE TABLE "log_metadata" (
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "finalizers" text NOT NULL DEFAULT '{}',
  "box_id" text NULL,
  "file_name" text NOT NULL,
  "format" text NOT NULL,
  "metadata" text NOT NULL,
  "total_line_bytes" bigint NOT NULL DEFAULT 0,
  "last_log_time" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "log_metadata_box_id_file_name_key" UNIQUE ("box_id", "file_name")
);
-- create "machine" table
CREATE TABLE "machine" (
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "finalizers" text NOT NULL DEFAULT '{}',
  "reconcile_status" text NOT NULL DEFAULT 'Initializing',
  "reconcile_status_details" text NOT NULL DEFAULT '',
  "name" text NOT NULL,
  "machine_provider_id" text NOT NULL,
  "machine_provider_type" text NOT NULL,
  "box_id" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "machine_workspace_id_name_key" UNIQUE ("workspace_id", "name")
);
-- create "machine_aws" table
CREATE TABLE "machine_aws" (
  "id" text NOT NULL,
  "reconcile_status" text NOT NULL DEFAULT 'Initializing',
  "reconcile_status_details" text NOT NULL DEFAULT '',
  "instance_type" text NOT NULL,
  "subnet_id" text NOT NULL,
  "root_volume_size" bigint NOT NULL,
  PRIMARY KEY ("id")
);
-- create "machine_aws_status" table
CREATE TABLE "machine_aws_status" (
  "id" text NOT NULL,
  "instance_id" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "machine_aws_status_instance_id_key" UNIQUE ("instance_id")
);
-- create "machine_hetzner" table
CREATE TABLE "machine_hetzner" (
  "id" text NOT NULL,
  "reconcile_status" text NOT NULL DEFAULT 'Initializing',
  "reconcile_status_details" text NOT NULL DEFAULT '',
  "server_type" text NOT NULL,
  "server_location" text NOT NULL,
  PRIMARY KEY ("id")
);
-- create "machine_hetzner_status" table
CREATE TABLE "machine_hetzner_status" (
  "id" text NOT NULL,
  "server_id" bigint NULL,
  PRIMARY KEY ("id")
);
-- create "machine_provider" table
CREATE TABLE "machine_provider" (
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
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
  "id" text NOT NULL,
  "region" text NOT NULL,
  "aws_access_key_id" text NULL,
  "aws_secret_access_key" text NULL,
  "vpc_id" text NOT NULL,
  PRIMARY KEY ("id")
);
-- create "machine_provider_aws_status" table
CREATE TABLE "machine_provider_aws_status" (
  "id" text NOT NULL,
  "vpc_name" text NULL,
  "vpc_cidr" text NULL,
  "security_group_id" text NULL,
  PRIMARY KEY ("id")
);
-- create "machine_provider_aws_subnet" table
CREATE TABLE "machine_provider_aws_subnet" (
  "machine_provider_id" text NOT NULL,
  "subnet_id" text NOT NULL,
  "subnet_name" text NULL,
  "availability_zone" text NOT NULL,
  "cidr" text NOT NULL,
  PRIMARY KEY ("machine_provider_id", "subnet_id")
);
-- create "machine_provider_hetzner" table
CREATE TABLE "machine_provider_hetzner" (
  "id" text NOT NULL,
  "hcloud_token" text NOT NULL,
  "hetzner_network_name" text NOT NULL,
  "robot_user" text NULL,
  "robot_password" text NULL,
  PRIMARY KEY ("id")
);
-- create "machine_provider_hetzner_status" table
CREATE TABLE "machine_provider_hetzner_status" (
  "id" text NOT NULL,
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
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
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
  "id" text NOT NULL,
  "netbird_version" text NOT NULL,
  "api_url" text NOT NULL,
  "api_access_token" text NOT NULL,
  PRIMARY KEY ("id")
);
-- create "s3_bucket" table
CREATE TABLE "s3_bucket" (
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "finalizers" text NOT NULL DEFAULT '{}',
  "reconcile_status" text NOT NULL DEFAULT 'Ok',
  "reconcile_status_details" text NOT NULL DEFAULT '',
  "endpoint" text NOT NULL,
  "bucket" text NOT NULL,
  "access_key_id" text NOT NULL,
  "secret_access_key" text NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "s3_bucket_workspace_bucket" to table: "s3_bucket"
CREATE INDEX "s3_bucket_workspace_bucket" ON "s3_bucket" ("workspace_id", "bucket");
-- create "token" table
CREATE TABLE "token" (
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "name" text NOT NULL,
  "token" text NOT NULL,
  "for_workspace" boolean NOT NULL,
  "box_id" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "token_token_key" UNIQUE ("token"),
  CONSTRAINT "token_workspace_id_name_key" UNIQUE ("workspace_id", "name")
);
-- create "user" table
CREATE TABLE "user" (
  "id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "username" text NULL,
  "email" text NULL,
  "full_name" text NULL,
  "avatar" text NULL,
  PRIMARY KEY ("id")
);
-- create "volume" table
CREATE TABLE "volume" (
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "finalizers" text NOT NULL DEFAULT '{}',
  "volume_provider_id" text NOT NULL,
  "volume_provider_type" text NOT NULL,
  "name" text NOT NULL,
  "lock_id" text NULL,
  "lock_time" timestamptz NULL,
  "lock_box_id" text NULL,
  "latest_snapshot_id" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "volume_workspace_id_name_key" UNIQUE ("workspace_id", "name")
);
-- create "volume_mount_status" table
CREATE TABLE "volume_mount_status" (
  "id" text NOT NULL,
  "status_time" timestamptz NULL,
  "run_status" text NULL,
  "start_time" timestamptz NULL,
  "stop_time" timestamptz NULL,
  "docker_ps" bytea NULL,
  PRIMARY KEY ("id")
);
-- create "volume_provider" table
CREATE TABLE "volume_provider" (
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
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
  "id" text NOT NULL,
  "storage_type" text NOT NULL,
  "s3_bucket_id" text NULL,
  "storage_prefix" text NOT NULL,
  "password" text NOT NULL,
  PRIMARY KEY ("id")
);
-- create "volume_rustic" table
CREATE TABLE "volume_rustic" (
  "id" text NOT NULL,
  "fs_size" bigint NOT NULL,
  "fs_type" text NOT NULL,
  PRIMARY KEY ("id")
);
-- create "volume_rustic_status" table
CREATE TABLE "volume_rustic_status" (
  "id" text NOT NULL,
  PRIMARY KEY ("id")
);
-- create "volume_snapshot" table
CREATE TABLE "volume_snapshot" (
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "finalizers" text NOT NULL DEFAULT '{}',
  "volume_provider_id" text NOT NULL,
  "volume_id" text NULL,
  "lock_id" text NOT NULL,
  PRIMARY KEY ("id")
);
-- create "volume_snapshot_rustic" table
CREATE TABLE "volume_snapshot_rustic" (
  "id" text NOT NULL,
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
  "id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "finalizers" text NOT NULL DEFAULT '{}',
  "reconcile_status" text NOT NULL DEFAULT 'Initializing',
  "reconcile_status_details" text NOT NULL DEFAULT '',
  "name" text NOT NULL,
  PRIMARY KEY ("id")
);
-- create "workspace_access" table
CREATE TABLE "workspace_access" (
  "workspace_id" text NOT NULL,
  "user_id" text NOT NULL,
  PRIMARY KEY ("workspace_id", "user_id")
);
-- create "workspace_quotas" table
CREATE TABLE "workspace_quotas" (
  "workspace_id" text NOT NULL,
  "max_log_bytes" integer NOT NULL DEFAULT 100,
  PRIMARY KEY ("workspace_id")
);
-- modify "box" table
ALTER TABLE "box" ADD CONSTRAINT "box_machine_id_fkey" FOREIGN KEY ("machine_id") REFERENCES "machine" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "box_network_id_fkey" FOREIGN KEY ("network_id") REFERENCES "network" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT, ADD CONSTRAINT "box_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- modify "box_compose_project" table
ALTER TABLE "box_compose_project" ADD CONSTRAINT "box_compose_project_box_id_fkey" FOREIGN KEY ("box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "box_netbird" table
ALTER TABLE "box_netbird" ADD CONSTRAINT "box_netbird_id_fkey" FOREIGN KEY ("id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "box_sandbox_status" table
ALTER TABLE "box_sandbox_status" ADD CONSTRAINT "box_sandbox_status_id_fkey" FOREIGN KEY ("id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "box_volume_attachment" table
ALTER TABLE "box_volume_attachment" ADD CONSTRAINT "box_volume_attachment_box_id_fkey" FOREIGN KEY ("box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "box_volume_attachment_volume_id_fkey" FOREIGN KEY ("volume_id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- modify "log_line" table
ALTER TABLE "log_line" ADD CONSTRAINT "log_line_log_id_fkey" FOREIGN KEY ("log_id") REFERENCES "log_metadata" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "log_line_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "log_metadata" table
ALTER TABLE "log_metadata" ADD CONSTRAINT "log_metadata_box_id_fkey" FOREIGN KEY ("box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "log_metadata_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
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
-- modify "s3_bucket" table
ALTER TABLE "s3_bucket" ADD CONSTRAINT "s3_bucket_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- modify "token" table
ALTER TABLE "token" ADD CONSTRAINT "token_box_id_fkey" FOREIGN KEY ("box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "token_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "volume" table
ALTER TABLE "volume" ADD CONSTRAINT "volume_latest_snapshot_id_fkey" FOREIGN KEY ("latest_snapshot_id") REFERENCES "volume_snapshot" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT, ADD CONSTRAINT "volume_lock_box_id_fkey" FOREIGN KEY ("lock_box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT, ADD CONSTRAINT "volume_volume_provider_id_fkey" FOREIGN KEY ("volume_provider_id") REFERENCES "volume_provider" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT, ADD CONSTRAINT "volume_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- modify "volume_mount_status" table
ALTER TABLE "volume_mount_status" ADD CONSTRAINT "volume_mount_status_id_fkey" FOREIGN KEY ("id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "volume_provider" table
ALTER TABLE "volume_provider" ADD CONSTRAINT "volume_provider_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- modify "volume_provider_rustic" table
ALTER TABLE "volume_provider_rustic" ADD CONSTRAINT "volume_provider_rustic_id_fkey" FOREIGN KEY ("id") REFERENCES "volume_provider" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "volume_provider_rustic_s3_bucket_id_fkey" FOREIGN KEY ("s3_bucket_id") REFERENCES "s3_bucket" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- modify "volume_rustic" table
ALTER TABLE "volume_rustic" ADD CONSTRAINT "volume_rustic_id_fkey" FOREIGN KEY ("id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "volume_rustic_status" table
ALTER TABLE "volume_rustic_status" ADD CONSTRAINT "volume_rustic_status_id_fkey" FOREIGN KEY ("id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "volume_snapshot" table
ALTER TABLE "volume_snapshot" ADD CONSTRAINT "volume_snapshot_volume_id_fkey" FOREIGN KEY ("volume_id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT, ADD CONSTRAINT "volume_snapshot_volume_provider_id_fkey" FOREIGN KEY ("volume_provider_id") REFERENCES "volume_provider" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT, ADD CONSTRAINT "volume_snapshot_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- modify "volume_snapshot_rustic" table
ALTER TABLE "volume_snapshot_rustic" ADD CONSTRAINT "volume_snapshot_rustic_id_fkey" FOREIGN KEY ("id") REFERENCES "volume_snapshot" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "workspace_access" table
ALTER TABLE "workspace_access" ADD CONSTRAINT "workspace_access_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "user" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT, ADD CONSTRAINT "workspace_access_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "workspace_quotas" table
ALTER TABLE "workspace_quotas" ADD CONSTRAINT "workspace_quotas_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
-- reverse: modify "workspace_quotas" table
ALTER TABLE "workspace_quotas" DROP CONSTRAINT "workspace_quotas_workspace_id_fkey";
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
-- reverse: modify "volume_provider_rustic" table
ALTER TABLE "volume_provider_rustic" DROP CONSTRAINT "volume_provider_rustic_s3_bucket_id_fkey", DROP CONSTRAINT "volume_provider_rustic_id_fkey";
-- reverse: modify "volume_provider" table
ALTER TABLE "volume_provider" DROP CONSTRAINT "volume_provider_workspace_id_fkey";
-- reverse: modify "volume_mount_status" table
ALTER TABLE "volume_mount_status" DROP CONSTRAINT "volume_mount_status_id_fkey";
-- reverse: modify "volume" table
ALTER TABLE "volume" DROP CONSTRAINT "volume_workspace_id_fkey", DROP CONSTRAINT "volume_volume_provider_id_fkey", DROP CONSTRAINT "volume_lock_box_id_fkey", DROP CONSTRAINT "volume_latest_snapshot_id_fkey";
-- reverse: modify "token" table
ALTER TABLE "token" DROP CONSTRAINT "token_workspace_id_fkey", DROP CONSTRAINT "token_box_id_fkey";
-- reverse: modify "s3_bucket" table
ALTER TABLE "s3_bucket" DROP CONSTRAINT "s3_bucket_workspace_id_fkey";
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
-- reverse: modify "log_metadata" table
ALTER TABLE "log_metadata" DROP CONSTRAINT "log_metadata_workspace_id_fkey", DROP CONSTRAINT "log_metadata_box_id_fkey";
-- reverse: modify "log_line" table
ALTER TABLE "log_line" DROP CONSTRAINT "log_line_workspace_id_fkey", DROP CONSTRAINT "log_line_log_id_fkey";
-- reverse: modify "box_volume_attachment" table
ALTER TABLE "box_volume_attachment" DROP CONSTRAINT "box_volume_attachment_volume_id_fkey", DROP CONSTRAINT "box_volume_attachment_box_id_fkey";
-- reverse: modify "box_sandbox_status" table
ALTER TABLE "box_sandbox_status" DROP CONSTRAINT "box_sandbox_status_id_fkey";
-- reverse: modify "box_netbird" table
ALTER TABLE "box_netbird" DROP CONSTRAINT "box_netbird_id_fkey";
-- reverse: modify "box_compose_project" table
ALTER TABLE "box_compose_project" DROP CONSTRAINT "box_compose_project_box_id_fkey";
-- reverse: modify "box" table
ALTER TABLE "box" DROP CONSTRAINT "box_workspace_id_fkey", DROP CONSTRAINT "box_network_id_fkey", DROP CONSTRAINT "box_machine_id_fkey";
-- reverse: create "workspace_quotas" table
DROP TABLE "workspace_quotas";
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
-- reverse: create "volume_provider_rustic" table
DROP TABLE "volume_provider_rustic";
-- reverse: create "volume_provider" table
DROP TABLE "volume_provider";
-- reverse: create "volume_mount_status" table
DROP TABLE "volume_mount_status";
-- reverse: create "volume" table
DROP TABLE "volume";
-- reverse: create "user" table
DROP TABLE "user";
-- reverse: create "token" table
DROP TABLE "token";
-- reverse: create index "s3_bucket_workspace_bucket" to table: "s3_bucket"
DROP INDEX "s3_bucket_workspace_bucket";
-- reverse: create "s3_bucket" table
DROP TABLE "s3_bucket";
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
-- reverse: create "log_metadata" table
DROP TABLE "log_metadata";
-- reverse: create index "log_line_time_index" to table: "log_line"
DROP INDEX "log_line_time_index";
-- reverse: create index "log_line_log_id_and_id" to table: "log_line"
DROP INDEX "log_line_log_id_and_id";
-- reverse: create "log_line" table
DROP TABLE "log_line";
-- reverse: create index "idx_table_and_id" to table: "change_tracking"
DROP INDEX "idx_table_and_id";
-- reverse: create "change_tracking" table
DROP TABLE "change_tracking";
-- reverse: create "box_volume_attachment" table
DROP TABLE "box_volume_attachment";
-- reverse: create "box_sandbox_status" table
DROP TABLE "box_sandbox_status";
-- reverse: create "box_netbird" table
DROP TABLE "box_netbird";
-- reverse: create "box_compose_project" table
DROP TABLE "box_compose_project";
-- reverse: create "box" table
DROP TABLE "box";
