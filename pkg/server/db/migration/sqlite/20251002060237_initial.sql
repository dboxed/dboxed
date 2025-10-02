-- +goose Up
-- create "change_tracking" table
CREATE TABLE `change_tracking` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `table_name` text NOT NULL,
  `entity_id` bigint NOT NULL,
  `time` datetime NOT NULL DEFAULT (current_timestamp)
);
-- create index "idx_table_and_id" to table: "change_tracking"
CREATE INDEX `idx_table_and_id` ON `change_tracking` (`table_name`, `id`);
-- create "user" table
CREATE TABLE `user` (
  `id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `name` text NOT NULL,
  `email` text NULL,
  `avatar` text NULL,
  PRIMARY KEY (`id`)
);
-- create "workspace" table
CREATE TABLE `workspace` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `name` text NOT NULL,
  `nkey` text NOT NULL,
  `nkey_seed` text NOT NULL
);
-- create index "workspace_nkey" to table: "workspace"
CREATE UNIQUE INDEX `workspace_nkey` ON `workspace` (`nkey`);
-- create "workspace_access" table
CREATE TABLE `workspace_access` (
  `workspace_id` bigint NOT NULL,
  `user_id` text NOT NULL,
  PRIMARY KEY (`workspace_id`, `user_id`),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `1` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "machine_provider" table
CREATE TABLE `machine_provider` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `workspace_id` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `type` text NOT NULL,
  `name` text NOT NULL,
  `ssh_key_public` text NULL,
  CONSTRAINT `0` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- create index "machine_provider_workspace_id_name" to table: "machine_provider"
CREATE UNIQUE INDEX `machine_provider_workspace_id_name` ON `machine_provider` (`workspace_id`, `name`);
-- create "machine_provider_aws" table
CREATE TABLE `machine_provider_aws` (
  `id` bigint NOT NULL,
  `region` text NOT NULL,
  `aws_access_key_id` text NULL,
  `aws_secret_access_key` text NULL,
  `vpc_id` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "machine_provider_aws_status" table
CREATE TABLE `machine_provider_aws_status` (
  `id` bigint NOT NULL,
  `vpc_name` text NULL,
  `vpc_cidr` text NULL,
  `security_group_id` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "machine_provider_aws_subnet" table
CREATE TABLE `machine_provider_aws_subnet` (
  `machine_provider_id` bigint NOT NULL,
  `subnet_id` text NOT NULL,
  `subnet_name` text NULL,
  `availability_zone` text NOT NULL,
  `cidr` text NOT NULL,
  PRIMARY KEY (`machine_provider_id`, `subnet_id`),
  CONSTRAINT `0` FOREIGN KEY (`machine_provider_id`) REFERENCES `machine_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "machine_provider_hetzner" table
CREATE TABLE `machine_provider_hetzner` (
  `id` bigint NOT NULL,
  `hcloud_token` text NOT NULL,
  `hetzner_network_name` text NOT NULL,
  `robot_user` text NULL,
  `robot_password` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "machine_provider_hetzner_status" table
CREATE TABLE `machine_provider_hetzner_status` (
  `id` bigint NOT NULL,
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
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `workspace_id` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `type` text NOT NULL,
  `name` text NOT NULL,
  CONSTRAINT `0` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- create index "network_workspace_id_name" to table: "network"
CREATE UNIQUE INDEX `network_workspace_id_name` ON `network` (`workspace_id`, `name`);
-- create "network_netbird" table
CREATE TABLE `network_netbird` (
  `id` bigint NOT NULL,
  `netbird_version` text NOT NULL,
  `api_url` text NOT NULL,
  `api_access_token` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `network` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "machine" table
CREATE TABLE `machine` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `workspace_id` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `name` text NOT NULL,
  `machine_provider_id` bigint NOT NULL,
  `machine_provider_type` text NOT NULL,
  `box_id` bigint NOT NULL,
  CONSTRAINT `0` FOREIGN KEY (`box_id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `1` FOREIGN KEY (`machine_provider_id`) REFERENCES `machine_provider` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `2` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- create index "machine_workspace_id_name" to table: "machine"
CREATE UNIQUE INDEX `machine_workspace_id_name` ON `machine` (`workspace_id`, `name`);
-- create "machine_aws" table
CREATE TABLE `machine_aws` (
  `id` bigint NOT NULL,
  `instance_type` text NOT NULL,
  `subnet_id` text NOT NULL,
  `root_volume_size` bigint NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "machine_aws_status" table
CREATE TABLE `machine_aws_status` (
  `id` bigint NOT NULL,
  `instance_id` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "machine_aws_status_instance_id" to table: "machine_aws_status"
CREATE UNIQUE INDEX `machine_aws_status_instance_id` ON `machine_aws_status` (`instance_id`);
-- create "machine_hetzner" table
CREATE TABLE `machine_hetzner` (
  `id` bigint NOT NULL,
  `server_type` text NOT NULL,
  `server_location` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "machine_hetzner_status" table
CREATE TABLE `machine_hetzner_status` (
  `id` bigint NOT NULL,
  `server_id` bigint NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "box" table
CREATE TABLE `box` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `workspace_id` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `uuid` text NOT NULL DEFAULT '',
  `name` text NOT NULL,
  `network_id` bigint NULL,
  `network_type` text NULL,
  `dboxed_version` text NOT NULL,
  `box_spec` bytea NOT NULL,
  `nkey` text NOT NULL,
  `nkey_seed` text NOT NULL,
  `machine_id` bigint NULL,
  CONSTRAINT `0` FOREIGN KEY (`machine_id`) REFERENCES `machine` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `1` FOREIGN KEY (`network_id`) REFERENCES `network` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `2` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- create index "box_uuid" to table: "box"
CREATE UNIQUE INDEX `box_uuid` ON `box` (`uuid`);
-- create index "box_nkey" to table: "box"
CREATE UNIQUE INDEX `box_nkey` ON `box` (`nkey`);
-- create index "box_workspace_id_name" to table: "box"
CREATE UNIQUE INDEX `box_workspace_id_name` ON `box` (`workspace_id`, `name`);
-- create "box_netbird" table
CREATE TABLE `box_netbird` (
  `id` bigint NOT NULL,
  `setup_key_id` text NULL,
  `setup_key` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "volume_provider" table
CREATE TABLE `volume_provider` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `workspace_id` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `type` text NOT NULL,
  `name` text NOT NULL,
  CONSTRAINT `0` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- create index "volume_provider_workspace_id_name" to table: "volume_provider"
CREATE UNIQUE INDEX `volume_provider_workspace_id_name` ON `volume_provider` (`workspace_id`, `name`);
-- create "volume_provider_rustic" table
CREATE TABLE `volume_provider_rustic` (
  `id` bigint NOT NULL,
  `storage_type` text NOT NULL,
  `password` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `volume_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "volume_provider_storage_s3" table
CREATE TABLE `volume_provider_storage_s3` (
  `id` bigint NOT NULL,
  `endpoint` text NOT NULL,
  `region` text NULL,
  `bucket` text NOT NULL,
  `access_key_id` text NOT NULL,
  `secret_access_key` text NOT NULL,
  `prefix` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `volume_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "volume" table
CREATE TABLE `volume` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `workspace_id` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `volume_provider_id` bigint NOT NULL,
  `volume_provider_type` text NOT NULL,
  `uuid` text NOT NULL,
  `name` text NOT NULL,
  `lock_id` text NULL,
  `lock_time` datetime NULL,
  `lock_box_uuid` text NULL,
  `latest_snapshot_id` bigint NULL,
  CONSTRAINT `0` FOREIGN KEY (`latest_snapshot_id`) REFERENCES `volume_snapshot` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `1` FOREIGN KEY (`volume_provider_id`) REFERENCES `volume_provider` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `2` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- create index "volume_uuid" to table: "volume"
CREATE UNIQUE INDEX `volume_uuid` ON `volume` (`uuid`);
-- create index "volume_workspace_id_name" to table: "volume"
CREATE UNIQUE INDEX `volume_workspace_id_name` ON `volume` (`workspace_id`, `name`);
-- create "volume_rustic" table
CREATE TABLE `volume_rustic` (
  `id` bigint NOT NULL,
  `fs_size` bigint NOT NULL,
  `fs_type` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `volume` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "volume_rustic_status" table
CREATE TABLE `volume_rustic_status` (
  `id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `volume` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "box_volume_attachment" table
CREATE TABLE `box_volume_attachment` (
  `box_id` bigint NOT NULL,
  `volume_id` bigint NOT NULL,
  `root_uid` bigint NOT NULL,
  `root_gid` bigint NOT NULL,
  `root_mode` text NOT NULL,
  PRIMARY KEY (`box_id`, `volume_id`),
  CONSTRAINT `0` FOREIGN KEY (`volume_id`) REFERENCES `volume` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `1` FOREIGN KEY (`box_id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "box_volume_attachment_volume_id" to table: "box_volume_attachment"
CREATE UNIQUE INDEX `box_volume_attachment_volume_id` ON `box_volume_attachment` (`volume_id`);
-- create "volume_snapshot" table
CREATE TABLE `volume_snapshot` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `workspace_id` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `volume_provider_id` bigint NOT NULL,
  `volume_id` bigint NULL,
  `lock_id` text NOT NULL,
  CONSTRAINT `0` FOREIGN KEY (`volume_id`) REFERENCES `volume` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `1` FOREIGN KEY (`volume_provider_id`) REFERENCES `volume_provider` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `2` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- create "volume_snapshot_rustic" table
CREATE TABLE `volume_snapshot_rustic` (
  `id` bigint NOT NULL,
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
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `workspace_id` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `name` text NOT NULL,
  `token` text NOT NULL,
  `for_workspace` bool NOT NULL,
  `box_id` bigint NULL,
  CONSTRAINT `0` FOREIGN KEY (`box_id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "token_token" to table: "token"
CREATE UNIQUE INDEX `token_token` ON `token` (`token`);
-- create index "token_workspace_id_name" to table: "token"
CREATE UNIQUE INDEX `token_workspace_id_name` ON `token` (`workspace_id`, `name`);

-- +goose Down
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
-- reverse: create index "volume_uuid" to table: "volume"
DROP INDEX `volume_uuid`;
-- reverse: create "volume" table
DROP TABLE `volume`;
-- reverse: create "volume_provider_storage_s3" table
DROP TABLE `volume_provider_storage_s3`;
-- reverse: create "volume_provider_rustic" table
DROP TABLE `volume_provider_rustic`;
-- reverse: create index "volume_provider_workspace_id_name" to table: "volume_provider"
DROP INDEX `volume_provider_workspace_id_name`;
-- reverse: create "volume_provider" table
DROP TABLE `volume_provider`;
-- reverse: create "box_netbird" table
DROP TABLE `box_netbird`;
-- reverse: create index "box_workspace_id_name" to table: "box"
DROP INDEX `box_workspace_id_name`;
-- reverse: create index "box_nkey" to table: "box"
DROP INDEX `box_nkey`;
-- reverse: create index "box_uuid" to table: "box"
DROP INDEX `box_uuid`;
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
-- reverse: create "workspace_access" table
DROP TABLE `workspace_access`;
-- reverse: create index "workspace_nkey" to table: "workspace"
DROP INDEX `workspace_nkey`;
-- reverse: create "workspace" table
DROP TABLE `workspace`;
-- reverse: create "user" table
DROP TABLE `user`;
-- reverse: create index "idx_table_and_id" to table: "change_tracking"
DROP INDEX `idx_table_and_id`;
-- reverse: create "change_tracking" table
DROP TABLE `change_tracking`;
