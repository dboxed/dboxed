-- +goose Up
-- create "change_tracking" table
CREATE TABLE `change_tracking` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `table_name` text NOT NULL,
  `entity_id` text NOT NULL,
  `time` datetime NOT NULL DEFAULT (current_timestamp)
);
-- create index "idx_table_and_id" to table: "change_tracking"
CREATE INDEX `idx_table_and_id` ON `change_tracking` (`table_name`, `id`);
-- create "user" table
CREATE TABLE `user` (
  `id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `username` text NULL,
  `email` text NULL,
  `full_name` text NULL,
  `avatar` text NULL,
  PRIMARY KEY (`id`)
);
-- create "workspace" table
CREATE TABLE `workspace` (
  `id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `name` text NOT NULL,
  PRIMARY KEY (`id`)
);
-- create "workspace_access" table
CREATE TABLE `workspace_access` (
  `workspace_id` text NOT NULL,
  `user_id` text NOT NULL,
  PRIMARY KEY (`workspace_id`, `user_id`),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `1` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "workspace_quotas" table
CREATE TABLE `workspace_quotas` (
  `workspace_id` text NOT NULL,
  `max_log_bytes` int NOT NULL DEFAULT 100,
  PRIMARY KEY (`workspace_id`),
  CONSTRAINT `0` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "machine_provider" table
CREATE TABLE `machine_provider` (
  `id` text NOT NULL,
  `workspace_id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `type` text NOT NULL,
  `name` text NOT NULL,
  `ssh_key_public` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- create index "machine_provider_workspace_id_name" to table: "machine_provider"
CREATE UNIQUE INDEX `machine_provider_workspace_id_name` ON `machine_provider` (`workspace_id`, `name`);
-- create "machine_provider_aws" table
CREATE TABLE `machine_provider_aws` (
  `id` text NOT NULL,
  `region` text NOT NULL,
  `aws_access_key_id` text NULL,
  `aws_secret_access_key` text NULL,
  `vpc_id` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "machine_provider_aws_status" table
CREATE TABLE `machine_provider_aws_status` (
  `id` text NOT NULL,
  `vpc_name` text NULL,
  `vpc_cidr` text NULL,
  `security_group_id` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "machine_provider_aws_subnet" table
CREATE TABLE `machine_provider_aws_subnet` (
  `machine_provider_id` text NOT NULL,
  `subnet_id` text NOT NULL,
  `subnet_name` text NULL,
  `availability_zone` text NOT NULL,
  `cidr` text NOT NULL,
  PRIMARY KEY (`machine_provider_id`, `subnet_id`),
  CONSTRAINT `0` FOREIGN KEY (`machine_provider_id`) REFERENCES `machine_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "machine_provider_hetzner" table
CREATE TABLE `machine_provider_hetzner` (
  `id` text NOT NULL,
  `hcloud_token` text NOT NULL,
  `hetzner_network_name` text NOT NULL,
  `robot_user` text NULL,
  `robot_password` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "machine_provider_hetzner_status" table
CREATE TABLE `machine_provider_hetzner_status` (
  `id` text NOT NULL,
  `hetzner_network_id` bigint NULL,
  `hetzner_network_zone` text NULL,
  `hetzner_network_cidr` text NULL,
  `cloud_subnet_cidr` text NULL,
  `robot_subnet_cidr` text NULL,
  `robot_vswitch_id` bigint NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "network" table
CREATE TABLE `network` (
  `id` text NOT NULL,
  `workspace_id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `type` text NOT NULL,
  `name` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- create index "network_workspace_id_name" to table: "network"
CREATE UNIQUE INDEX `network_workspace_id_name` ON `network` (`workspace_id`, `name`);
-- create "network_netbird" table
CREATE TABLE `network_netbird` (
  `id` text NOT NULL,
  `netbird_version` text NOT NULL,
  `api_url` text NOT NULL,
  `api_access_token` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `network` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "machine" table
CREATE TABLE `machine` (
  `id` text NOT NULL,
  `workspace_id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `name` text NOT NULL,
  `machine_provider_id` text NOT NULL,
  `machine_provider_type` text NOT NULL,
  `box_id` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`box_id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `1` FOREIGN KEY (`machine_provider_id`) REFERENCES `machine_provider` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `2` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- create index "machine_workspace_id_name" to table: "machine"
CREATE UNIQUE INDEX `machine_workspace_id_name` ON `machine` (`workspace_id`, `name`);
-- create "machine_aws" table
CREATE TABLE `machine_aws` (
  `id` text NOT NULL,
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `instance_type` text NOT NULL,
  `subnet_id` text NOT NULL,
  `root_volume_size` bigint NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "machine_aws_status" table
CREATE TABLE `machine_aws_status` (
  `id` text NOT NULL,
  `instance_id` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "machine_aws_status_instance_id" to table: "machine_aws_status"
CREATE UNIQUE INDEX `machine_aws_status_instance_id` ON `machine_aws_status` (`instance_id`);
-- create "machine_hetzner" table
CREATE TABLE `machine_hetzner` (
  `id` text NOT NULL,
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `server_type` text NOT NULL,
  `server_location` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "machine_hetzner_status" table
CREATE TABLE `machine_hetzner_status` (
  `id` text NOT NULL,
  `server_id` bigint NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "box" table
CREATE TABLE `box` (
  `id` text NOT NULL,
  `workspace_id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `name` text NOT NULL,
  `network_id` text NULL,
  `network_type` text NULL,
  `dboxed_version` text NOT NULL,
  `machine_id` text NULL,
  `desired_state` text NOT NULL DEFAULT 'up',
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`machine_id`) REFERENCES `machine` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `1` FOREIGN KEY (`network_id`) REFERENCES `network` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `2` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- create index "box_workspace_id_name" to table: "box"
CREATE UNIQUE INDEX `box_workspace_id_name` ON `box` (`workspace_id`, `name`);
-- create "box_sandbox_status" table
CREATE TABLE `box_sandbox_status` (
  `id` text NOT NULL,
  `status_time` datetime NULL,
  `run_status` text NULL,
  `start_time` datetime NULL,
  `stop_time` datetime NULL,
  `docker_ps` blob NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "box_netbird" table
CREATE TABLE `box_netbird` (
  `id` text NOT NULL,
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `setup_key_id` text NULL,
  `setup_key` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "box_compose_project" table
CREATE TABLE `box_compose_project` (
  `box_id` text NOT NULL,
  `name` text NOT NULL,
  `compose_project` text NOT NULL,
  PRIMARY KEY (`box_id`, `name`),
  CONSTRAINT `0` FOREIGN KEY (`box_id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "s3_bucket" table
CREATE TABLE `s3_bucket` (
  `id` text NOT NULL,
  `workspace_id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `reconcile_status` text NOT NULL DEFAULT 'Ok',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `endpoint` text NOT NULL,
  `bucket` text NOT NULL,
  `access_key_id` text NOT NULL,
  `secret_access_key` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- create index "s3_bucket_workspace_bucket" to table: "s3_bucket"
CREATE INDEX `s3_bucket_workspace_bucket` ON `s3_bucket` (`workspace_id`, `bucket`);
-- create "volume_provider" table
CREATE TABLE `volume_provider` (
  `id` text NOT NULL,
  `workspace_id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `type` text NOT NULL,
  `name` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- create index "volume_provider_workspace_id_name" to table: "volume_provider"
CREATE UNIQUE INDEX `volume_provider_workspace_id_name` ON `volume_provider` (`workspace_id`, `name`);
-- create "volume_provider_rustic" table
CREATE TABLE `volume_provider_rustic` (
  `id` text NOT NULL,
  `storage_type` text NOT NULL,
  `s3_bucket_id` text NULL,
  `storage_prefix` text NOT NULL,
  `password` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`s3_bucket_id`) REFERENCES `s3_bucket` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `1` FOREIGN KEY (`id`) REFERENCES `volume_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "volume" table
CREATE TABLE `volume` (
  `id` text NOT NULL,
  `workspace_id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `volume_provider_id` text NOT NULL,
  `volume_provider_type` text NOT NULL,
  `name` text NOT NULL,
  `lock_id` text NULL,
  `lock_time` datetime NULL,
  `lock_box_id` text NULL,
  `latest_snapshot_id` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`latest_snapshot_id`) REFERENCES `volume_snapshot` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `1` FOREIGN KEY (`lock_box_id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `2` FOREIGN KEY (`volume_provider_id`) REFERENCES `volume_provider` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `3` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- create index "volume_workspace_id_name" to table: "volume"
CREATE UNIQUE INDEX `volume_workspace_id_name` ON `volume` (`workspace_id`, `name`);
-- create "volume_rustic" table
CREATE TABLE `volume_rustic` (
  `id` text NOT NULL,
  `fs_size` bigint NOT NULL,
  `fs_type` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `volume` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "volume_rustic_status" table
CREATE TABLE `volume_rustic_status` (
  `id` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `volume` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "box_volume_attachment" table
CREATE TABLE `box_volume_attachment` (
  `box_id` text NOT NULL,
  `volume_id` text NOT NULL,
  `root_uid` bigint NOT NULL,
  `root_gid` bigint NOT NULL,
  `root_mode` text NOT NULL,
  PRIMARY KEY (`box_id`, `volume_id`),
  CONSTRAINT `0` FOREIGN KEY (`volume_id`) REFERENCES `volume` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `1` FOREIGN KEY (`box_id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "box_volume_attachment_volume_id" to table: "box_volume_attachment"
CREATE UNIQUE INDEX `box_volume_attachment_volume_id` ON `box_volume_attachment` (`volume_id`);
-- create "volume_mount_status" table
CREATE TABLE `volume_mount_status` (
  `id` text NOT NULL,
  `status_time` datetime NULL,
  `run_status` text NULL,
  `start_time` datetime NULL,
  `stop_time` datetime NULL,
  `docker_ps` blob NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `volume` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "volume_snapshot" table
CREATE TABLE `volume_snapshot` (
  `id` text NOT NULL,
  `workspace_id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `volume_provider_id` text NOT NULL,
  `volume_id` text NULL,
  `lock_id` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`volume_id`) REFERENCES `volume` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `1` FOREIGN KEY (`volume_provider_id`) REFERENCES `volume_provider` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `2` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- create "volume_snapshot_rustic" table
CREATE TABLE `volume_snapshot_rustic` (
  `id` text NOT NULL,
  `snapshot_id` text NOT NULL,
  `snapshot_time` datetime NOT NULL,
  `parent_snapshot_id` text NULL,
  `hostname` text NOT NULL,
  `files_new` int NOT NULL,
  `files_changed` int NOT NULL,
  `files_unmodified` int NOT NULL,
  `total_files_processed` int NOT NULL,
  `total_bytes_processed` int NOT NULL,
  `dirs_new` int NOT NULL,
  `dirs_changed` int NOT NULL,
  `dirs_unmodified` int NOT NULL,
  `total_dirs_processed` int NOT NULL,
  `total_dirsize_processed` int NOT NULL,
  `data_blobs` int NOT NULL,
  `tree_blobs` int NOT NULL,
  `data_added` int NOT NULL,
  `data_added_packed` int NOT NULL,
  `data_added_files` int NOT NULL,
  `data_added_files_packed` int NOT NULL,
  `data_added_trees` int NOT NULL,
  `data_added_trees_packed` int NOT NULL,
  `backup_start` datetime NOT NULL,
  `backup_end` datetime NOT NULL,
  `backup_duration` float4 NOT NULL,
  `total_duration` float4 NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `volume_snapshot` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "volume_snapshot_rustic_snapshot_id" to table: "volume_snapshot_rustic"
CREATE UNIQUE INDEX `volume_snapshot_rustic_snapshot_id` ON `volume_snapshot_rustic` (`snapshot_id`);
-- create "token" table
CREATE TABLE `token` (
  `id` text NOT NULL,
  `workspace_id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `name` text NOT NULL,
  `token` text NOT NULL,
  `for_workspace` bool NOT NULL,
  `box_id` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`box_id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "token_token" to table: "token"
CREATE UNIQUE INDEX `token_token` ON `token` (`token`);
-- create index "token_workspace_id_name" to table: "token"
CREATE UNIQUE INDEX `token_workspace_id_name` ON `token` (`workspace_id`, `name`);
-- create "log_metadata" table
CREATE TABLE `log_metadata` (
  `id` text NOT NULL,
  `workspace_id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `box_id` text NULL,
  `file_name` text NOT NULL,
  `format` text NOT NULL,
  `metadata` text NOT NULL,
  `total_line_bytes` bigint NOT NULL DEFAULT 0,
  `last_log_time` datetime NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`box_id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "log_metadata_box_id_file_name" to table: "log_metadata"
CREATE UNIQUE INDEX `log_metadata_box_id_file_name` ON `log_metadata` (`box_id`, `file_name`);
-- create "log_line" table
CREATE TABLE `log_line` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `workspace_id` text NOT NULL,
  `log_id` text NOT NULL,
  `time` datetime NOT NULL,
  `line` text NOT NULL,
  CONSTRAINT `0` FOREIGN KEY (`log_id`) REFERENCES `log_metadata` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "log_line_log_id_and_id" to table: "log_line"
CREATE INDEX `log_line_log_id_and_id` ON `log_line` (`log_id`, `id`);
-- create index "log_line_time_index" to table: "log_line"
CREATE INDEX `log_line_time_index` ON `log_line` (`log_id`, `time`);

-- +goose Down
-- reverse: create index "log_line_time_index" to table: "log_line"
DROP INDEX `log_line_time_index`;
-- reverse: create index "log_line_log_id_and_id" to table: "log_line"
DROP INDEX `log_line_log_id_and_id`;
-- reverse: create "log_line" table
DROP TABLE `log_line`;
-- reverse: create index "log_metadata_box_id_file_name" to table: "log_metadata"
DROP INDEX `log_metadata_box_id_file_name`;
-- reverse: create "log_metadata" table
DROP TABLE `log_metadata`;
-- reverse: create index "token_workspace_id_name" to table: "token"
DROP INDEX `token_workspace_id_name`;
-- reverse: create index "token_token" to table: "token"
DROP INDEX `token_token`;
-- reverse: create "token" table
DROP TABLE `token`;
-- reverse: create index "volume_snapshot_rustic_snapshot_id" to table: "volume_snapshot_rustic"
DROP INDEX `volume_snapshot_rustic_snapshot_id`;
-- reverse: create "volume_snapshot_rustic" table
DROP TABLE `volume_snapshot_rustic`;
-- reverse: create "volume_snapshot" table
DROP TABLE `volume_snapshot`;
-- reverse: create "volume_mount_status" table
DROP TABLE `volume_mount_status`;
-- reverse: create index "box_volume_attachment_volume_id" to table: "box_volume_attachment"
DROP INDEX `box_volume_attachment_volume_id`;
-- reverse: create "box_volume_attachment" table
DROP TABLE `box_volume_attachment`;
-- reverse: create "volume_rustic_status" table
DROP TABLE `volume_rustic_status`;
-- reverse: create "volume_rustic" table
DROP TABLE `volume_rustic`;
-- reverse: create index "volume_workspace_id_name" to table: "volume"
DROP INDEX `volume_workspace_id_name`;
-- reverse: create "volume" table
DROP TABLE `volume`;
-- reverse: create "volume_provider_rustic" table
DROP TABLE `volume_provider_rustic`;
-- reverse: create index "volume_provider_workspace_id_name" to table: "volume_provider"
DROP INDEX `volume_provider_workspace_id_name`;
-- reverse: create "volume_provider" table
DROP TABLE `volume_provider`;
-- reverse: create index "s3_bucket_workspace_bucket" to table: "s3_bucket"
DROP INDEX `s3_bucket_workspace_bucket`;
-- reverse: create "s3_bucket" table
DROP TABLE `s3_bucket`;
-- reverse: create "box_compose_project" table
DROP TABLE `box_compose_project`;
-- reverse: create "box_netbird" table
DROP TABLE `box_netbird`;
-- reverse: create "box_sandbox_status" table
DROP TABLE `box_sandbox_status`;
-- reverse: create index "box_workspace_id_name" to table: "box"
DROP INDEX `box_workspace_id_name`;
-- reverse: create "box" table
DROP TABLE `box`;
-- reverse: create "machine_hetzner_status" table
DROP TABLE `machine_hetzner_status`;
-- reverse: create "machine_hetzner" table
DROP TABLE `machine_hetzner`;
-- reverse: create index "machine_aws_status_instance_id" to table: "machine_aws_status"
DROP INDEX `machine_aws_status_instance_id`;
-- reverse: create "machine_aws_status" table
DROP TABLE `machine_aws_status`;
-- reverse: create "machine_aws" table
DROP TABLE `machine_aws`;
-- reverse: create index "machine_workspace_id_name" to table: "machine"
DROP INDEX `machine_workspace_id_name`;
-- reverse: create "machine" table
DROP TABLE `machine`;
-- reverse: create "network_netbird" table
DROP TABLE `network_netbird`;
-- reverse: create index "network_workspace_id_name" to table: "network"
DROP INDEX `network_workspace_id_name`;
-- reverse: create "network" table
DROP TABLE `network`;
-- reverse: create "machine_provider_hetzner_status" table
DROP TABLE `machine_provider_hetzner_status`;
-- reverse: create "machine_provider_hetzner" table
DROP TABLE `machine_provider_hetzner`;
-- reverse: create "machine_provider_aws_subnet" table
DROP TABLE `machine_provider_aws_subnet`;
-- reverse: create "machine_provider_aws_status" table
DROP TABLE `machine_provider_aws_status`;
-- reverse: create "machine_provider_aws" table
DROP TABLE `machine_provider_aws`;
-- reverse: create index "machine_provider_workspace_id_name" to table: "machine_provider"
DROP INDEX `machine_provider_workspace_id_name`;
-- reverse: create "machine_provider" table
DROP TABLE `machine_provider`;
-- reverse: create "workspace_quotas" table
DROP TABLE `workspace_quotas`;
-- reverse: create "workspace_access" table
DROP TABLE `workspace_access`;
-- reverse: create "workspace" table
DROP TABLE `workspace`;
-- reverse: create "user" table
DROP TABLE `user`;
-- reverse: create index "idx_table_and_id" to table: "change_tracking"
DROP INDEX `idx_table_and_id`;
-- reverse: create "change_tracking" table
DROP TABLE `change_tracking`;
