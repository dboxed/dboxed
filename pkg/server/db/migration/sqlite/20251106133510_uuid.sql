-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_change_tracking" table
CREATE TABLE `new_change_tracking` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `table_name` text NOT NULL,
  `entity_id` text NOT NULL,
  `time` datetime NOT NULL DEFAULT (current_timestamp)
);
-- copy rows from old table "change_tracking" to new temporary table "new_change_tracking"
INSERT INTO `new_change_tracking` (`id`, `table_name`, `entity_id`, `time`) SELECT `id`, `table_name`, `entity_id`, `time` FROM `change_tracking`;
-- drop "change_tracking" table after copying rows
DROP TABLE `change_tracking`;
-- rename temporary table "new_change_tracking" to "change_tracking"
ALTER TABLE `new_change_tracking` RENAME TO `change_tracking`;
-- create index "idx_table_and_id" to table: "change_tracking"
CREATE INDEX `idx_table_and_id` ON `change_tracking` (`table_name`, `id`);
-- create "new_workspace" table
CREATE TABLE `new_workspace` (
  `id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `name` text NOT NULL,
  PRIMARY KEY (`id`)
);
-- copy rows from old table "workspace" to new temporary table "new_workspace"
INSERT INTO `new_workspace` (`id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `name`) SELECT `id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `name` FROM `workspace`;
-- drop "workspace" table after copying rows
DROP TABLE `workspace`;
-- rename temporary table "new_workspace" to "workspace"
ALTER TABLE `new_workspace` RENAME TO `workspace`;
-- create "new_workspace_access" table
CREATE TABLE `new_workspace_access` (
  `workspace_id` text NOT NULL,
  `user_id` text NOT NULL,
  PRIMARY KEY (`workspace_id`, `user_id`),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `1` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "workspace_access" to new temporary table "new_workspace_access"
INSERT INTO `new_workspace_access` (`workspace_id`, `user_id`) SELECT `workspace_id`, `user_id` FROM `workspace_access`;
-- drop "workspace_access" table after copying rows
DROP TABLE `workspace_access`;
-- rename temporary table "new_workspace_access" to "workspace_access"
ALTER TABLE `new_workspace_access` RENAME TO `workspace_access`;
-- create "new_workspace_quotas" table
CREATE TABLE `new_workspace_quotas` (
  `workspace_id` text NOT NULL,
  `max_log_bytes` int NOT NULL DEFAULT 100,
  PRIMARY KEY (`workspace_id`),
  CONSTRAINT `0` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "workspace_quotas" to new temporary table "new_workspace_quotas"
INSERT INTO `new_workspace_quotas` (`workspace_id`, `max_log_bytes`) SELECT `workspace_id`, `max_log_bytes` FROM `workspace_quotas`;
-- drop "workspace_quotas" table after copying rows
DROP TABLE `workspace_quotas`;
-- rename temporary table "new_workspace_quotas" to "workspace_quotas"
ALTER TABLE `new_workspace_quotas` RENAME TO `workspace_quotas`;
-- create "new_machine_provider" table
CREATE TABLE `new_machine_provider` (
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
-- copy rows from old table "machine_provider" to new temporary table "new_machine_provider"
INSERT INTO `new_machine_provider` (`id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `type`, `name`, `ssh_key_public`) SELECT `id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `type`, `name`, `ssh_key_public` FROM `machine_provider`;
-- drop "machine_provider" table after copying rows
DROP TABLE `machine_provider`;
-- rename temporary table "new_machine_provider" to "machine_provider"
ALTER TABLE `new_machine_provider` RENAME TO `machine_provider`;
-- create index "machine_provider_workspace_id_name" to table: "machine_provider"
CREATE UNIQUE INDEX `machine_provider_workspace_id_name` ON `machine_provider` (`workspace_id`, `name`);
-- create "new_machine_provider_aws" table
CREATE TABLE `new_machine_provider_aws` (
  `id` text NOT NULL,
  `region` text NOT NULL,
  `aws_access_key_id` text NULL,
  `aws_secret_access_key` text NULL,
  `vpc_id` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "machine_provider_aws" to new temporary table "new_machine_provider_aws"
INSERT INTO `new_machine_provider_aws` (`id`, `region`, `aws_access_key_id`, `aws_secret_access_key`, `vpc_id`) SELECT `id`, `region`, `aws_access_key_id`, `aws_secret_access_key`, `vpc_id` FROM `machine_provider_aws`;
-- drop "machine_provider_aws" table after copying rows
DROP TABLE `machine_provider_aws`;
-- rename temporary table "new_machine_provider_aws" to "machine_provider_aws"
ALTER TABLE `new_machine_provider_aws` RENAME TO `machine_provider_aws`;
-- create "new_machine_provider_aws_status" table
CREATE TABLE `new_machine_provider_aws_status` (
  `id` text NOT NULL,
  `vpc_name` text NULL,
  `vpc_cidr` text NULL,
  `security_group_id` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "machine_provider_aws_status" to new temporary table "new_machine_provider_aws_status"
INSERT INTO `new_machine_provider_aws_status` (`id`, `vpc_name`, `vpc_cidr`, `security_group_id`) SELECT `id`, `vpc_name`, `vpc_cidr`, `security_group_id` FROM `machine_provider_aws_status`;
-- drop "machine_provider_aws_status" table after copying rows
DROP TABLE `machine_provider_aws_status`;
-- rename temporary table "new_machine_provider_aws_status" to "machine_provider_aws_status"
ALTER TABLE `new_machine_provider_aws_status` RENAME TO `machine_provider_aws_status`;
-- create "new_machine_provider_aws_subnet" table
CREATE TABLE `new_machine_provider_aws_subnet` (
  `machine_provider_id` text NOT NULL,
  `subnet_id` text NOT NULL,
  `subnet_name` text NULL,
  `availability_zone` text NOT NULL,
  `cidr` text NOT NULL,
  PRIMARY KEY (`machine_provider_id`, `subnet_id`),
  CONSTRAINT `0` FOREIGN KEY (`machine_provider_id`) REFERENCES `machine_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "machine_provider_aws_subnet" to new temporary table "new_machine_provider_aws_subnet"
INSERT INTO `new_machine_provider_aws_subnet` (`machine_provider_id`, `subnet_id`, `subnet_name`, `availability_zone`, `cidr`) SELECT `machine_provider_id`, `subnet_id`, `subnet_name`, `availability_zone`, `cidr` FROM `machine_provider_aws_subnet`;
-- drop "machine_provider_aws_subnet" table after copying rows
DROP TABLE `machine_provider_aws_subnet`;
-- rename temporary table "new_machine_provider_aws_subnet" to "machine_provider_aws_subnet"
ALTER TABLE `new_machine_provider_aws_subnet` RENAME TO `machine_provider_aws_subnet`;
-- create "new_machine_provider_hetzner" table
CREATE TABLE `new_machine_provider_hetzner` (
  `id` text NOT NULL,
  `hcloud_token` text NOT NULL,
  `hetzner_network_name` text NOT NULL,
  `robot_user` text NULL,
  `robot_password` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "machine_provider_hetzner" to new temporary table "new_machine_provider_hetzner"
INSERT INTO `new_machine_provider_hetzner` (`id`, `hcloud_token`, `hetzner_network_name`, `robot_user`, `robot_password`) SELECT `id`, `hcloud_token`, `hetzner_network_name`, `robot_user`, `robot_password` FROM `machine_provider_hetzner`;
-- drop "machine_provider_hetzner" table after copying rows
DROP TABLE `machine_provider_hetzner`;
-- rename temporary table "new_machine_provider_hetzner" to "machine_provider_hetzner"
ALTER TABLE `new_machine_provider_hetzner` RENAME TO `machine_provider_hetzner`;
-- create "new_machine_provider_hetzner_status" table
CREATE TABLE `new_machine_provider_hetzner_status` (
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
-- copy rows from old table "machine_provider_hetzner_status" to new temporary table "new_machine_provider_hetzner_status"
INSERT INTO `new_machine_provider_hetzner_status` (`id`, `hetzner_network_id`, `hetzner_network_zone`, `hetzner_network_cidr`, `cloud_subnet_cidr`, `robot_subnet_cidr`, `robot_vswitch_id`) SELECT `id`, `hetzner_network_id`, `hetzner_network_zone`, `hetzner_network_cidr`, `cloud_subnet_cidr`, `robot_subnet_cidr`, `robot_vswitch_id` FROM `machine_provider_hetzner_status`;
-- drop "machine_provider_hetzner_status" table after copying rows
DROP TABLE `machine_provider_hetzner_status`;
-- rename temporary table "new_machine_provider_hetzner_status" to "machine_provider_hetzner_status"
ALTER TABLE `new_machine_provider_hetzner_status` RENAME TO `machine_provider_hetzner_status`;
-- create "new_network" table
CREATE TABLE `new_network` (
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
-- copy rows from old table "network" to new temporary table "new_network"
INSERT INTO `new_network` (`id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `type`, `name`) SELECT `id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `type`, `name` FROM `network`;
-- drop "network" table after copying rows
DROP TABLE `network`;
-- rename temporary table "new_network" to "network"
ALTER TABLE `new_network` RENAME TO `network`;
-- create index "network_workspace_id_name" to table: "network"
CREATE UNIQUE INDEX `network_workspace_id_name` ON `network` (`workspace_id`, `name`);
-- create "new_network_netbird" table
CREATE TABLE `new_network_netbird` (
  `id` text NOT NULL,
  `netbird_version` text NOT NULL,
  `api_url` text NOT NULL,
  `api_access_token` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `network` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "network_netbird" to new temporary table "new_network_netbird"
INSERT INTO `new_network_netbird` (`id`, `netbird_version`, `api_url`, `api_access_token`) SELECT `id`, `netbird_version`, `api_url`, `api_access_token` FROM `network_netbird`;
-- drop "network_netbird" table after copying rows
DROP TABLE `network_netbird`;
-- rename temporary table "new_network_netbird" to "network_netbird"
ALTER TABLE `new_network_netbird` RENAME TO `network_netbird`;
-- create "new_machine" table
CREATE TABLE `new_machine` (
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
-- copy rows from old table "machine" to new temporary table "new_machine"
INSERT INTO `new_machine` (`id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `name`, `machine_provider_id`, `machine_provider_type`, `box_id`) SELECT `id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `name`, `machine_provider_id`, `machine_provider_type`, `box_id` FROM `machine`;
-- drop "machine" table after copying rows
DROP TABLE `machine`;
-- rename temporary table "new_machine" to "machine"
ALTER TABLE `new_machine` RENAME TO `machine`;
-- create index "machine_workspace_id_name" to table: "machine"
CREATE UNIQUE INDEX `machine_workspace_id_name` ON `machine` (`workspace_id`, `name`);
-- create "new_machine_aws" table
CREATE TABLE `new_machine_aws` (
  `id` text NOT NULL,
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `instance_type` text NOT NULL,
  `subnet_id` text NOT NULL,
  `root_volume_size` bigint NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "machine_aws" to new temporary table "new_machine_aws"
INSERT INTO `new_machine_aws` (`id`, `reconcile_status`, `reconcile_status_details`, `instance_type`, `subnet_id`, `root_volume_size`) SELECT `id`, `reconcile_status`, `reconcile_status_details`, `instance_type`, `subnet_id`, `root_volume_size` FROM `machine_aws`;
-- drop "machine_aws" table after copying rows
DROP TABLE `machine_aws`;
-- rename temporary table "new_machine_aws" to "machine_aws"
ALTER TABLE `new_machine_aws` RENAME TO `machine_aws`;
-- create "new_machine_aws_status" table
CREATE TABLE `new_machine_aws_status` (
  `id` text NOT NULL,
  `instance_id` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "machine_aws_status" to new temporary table "new_machine_aws_status"
INSERT INTO `new_machine_aws_status` (`id`, `instance_id`) SELECT `id`, `instance_id` FROM `machine_aws_status`;
-- drop "machine_aws_status" table after copying rows
DROP TABLE `machine_aws_status`;
-- rename temporary table "new_machine_aws_status" to "machine_aws_status"
ALTER TABLE `new_machine_aws_status` RENAME TO `machine_aws_status`;
-- create index "machine_aws_status_instance_id" to table: "machine_aws_status"
CREATE UNIQUE INDEX `machine_aws_status_instance_id` ON `machine_aws_status` (`instance_id`);
-- create "new_machine_hetzner" table
CREATE TABLE `new_machine_hetzner` (
  `id` text NOT NULL,
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `server_type` text NOT NULL,
  `server_location` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "machine_hetzner" to new temporary table "new_machine_hetzner"
INSERT INTO `new_machine_hetzner` (`id`, `reconcile_status`, `reconcile_status_details`, `server_type`, `server_location`) SELECT `id`, `reconcile_status`, `reconcile_status_details`, `server_type`, `server_location` FROM `machine_hetzner`;
-- drop "machine_hetzner" table after copying rows
DROP TABLE `machine_hetzner`;
-- rename temporary table "new_machine_hetzner" to "machine_hetzner"
ALTER TABLE `new_machine_hetzner` RENAME TO `machine_hetzner`;
-- create "new_machine_hetzner_status" table
CREATE TABLE `new_machine_hetzner_status` (
  `id` text NOT NULL,
  `server_id` bigint NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `machine` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "machine_hetzner_status" to new temporary table "new_machine_hetzner_status"
INSERT INTO `new_machine_hetzner_status` (`id`, `server_id`) SELECT `id`, `server_id` FROM `machine_hetzner_status`;
-- drop "machine_hetzner_status" table after copying rows
DROP TABLE `machine_hetzner_status`;
-- rename temporary table "new_machine_hetzner_status" to "machine_hetzner_status"
ALTER TABLE `new_machine_hetzner_status` RENAME TO `machine_hetzner_status`;
-- create "new_box_netbird" table
CREATE TABLE `new_box_netbird` (
  `id` text NOT NULL,
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `setup_key_id` text NULL,
  `setup_key` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "box_netbird" to new temporary table "new_box_netbird"
INSERT INTO `new_box_netbird` (`id`, `reconcile_status`, `reconcile_status_details`, `setup_key_id`, `setup_key`) SELECT `id`, `reconcile_status`, `reconcile_status_details`, `setup_key_id`, `setup_key` FROM `box_netbird`;
-- drop "box_netbird" table after copying rows
DROP TABLE `box_netbird`;
-- rename temporary table "new_box_netbird" to "box_netbird"
ALTER TABLE `new_box_netbird` RENAME TO `box_netbird`;
-- create "new_box_compose_project" table
CREATE TABLE `new_box_compose_project` (
  `box_id` text NOT NULL,
  `name` text NOT NULL,
  `compose_project` text NOT NULL,
  PRIMARY KEY (`box_id`, `name`),
  CONSTRAINT `0` FOREIGN KEY (`box_id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "box_compose_project" to new temporary table "new_box_compose_project"
INSERT INTO `new_box_compose_project` (`box_id`, `name`, `compose_project`) SELECT `box_id`, `name`, `compose_project` FROM `box_compose_project`;
-- drop "box_compose_project" table after copying rows
DROP TABLE `box_compose_project`;
-- rename temporary table "new_box_compose_project" to "box_compose_project"
ALTER TABLE `new_box_compose_project` RENAME TO `box_compose_project`;
-- create "new_volume_provider" table
CREATE TABLE `new_volume_provider` (
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
-- copy rows from old table "volume_provider" to new temporary table "new_volume_provider"
INSERT INTO `new_volume_provider` (`id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `type`, `name`) SELECT `id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `type`, `name` FROM `volume_provider`;
-- drop "volume_provider" table after copying rows
DROP TABLE `volume_provider`;
-- rename temporary table "new_volume_provider" to "volume_provider"
ALTER TABLE `new_volume_provider` RENAME TO `volume_provider`;
-- create index "volume_provider_workspace_id_name" to table: "volume_provider"
CREATE UNIQUE INDEX `volume_provider_workspace_id_name` ON `volume_provider` (`workspace_id`, `name`);
-- create "new_volume" table
CREATE TABLE `new_volume` (
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
-- copy rows from old table "volume" to new temporary table "new_volume"
INSERT INTO `new_volume` (`id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `volume_provider_id`, `volume_provider_type`, `name`, `lock_id`, `lock_time`, `lock_box_id`, `latest_snapshot_id`) SELECT `id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `volume_provider_id`, `volume_provider_type`, `name`, `lock_id`, `lock_time`, `lock_box_id`, `latest_snapshot_id` FROM `volume`;
-- drop "volume" table after copying rows
DROP TABLE `volume`;
-- rename temporary table "new_volume" to "volume"
ALTER TABLE `new_volume` RENAME TO `volume`;
-- create index "volume_workspace_id_name" to table: "volume"
CREATE UNIQUE INDEX `volume_workspace_id_name` ON `volume` (`workspace_id`, `name`);
-- create "new_volume_rustic" table
CREATE TABLE `new_volume_rustic` (
  `id` text NOT NULL,
  `fs_size` bigint NOT NULL,
  `fs_type` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `volume` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "volume_rustic" to new temporary table "new_volume_rustic"
INSERT INTO `new_volume_rustic` (`id`, `fs_size`, `fs_type`) SELECT `id`, `fs_size`, `fs_type` FROM `volume_rustic`;
-- drop "volume_rustic" table after copying rows
DROP TABLE `volume_rustic`;
-- rename temporary table "new_volume_rustic" to "volume_rustic"
ALTER TABLE `new_volume_rustic` RENAME TO `volume_rustic`;
-- create "new_volume_rustic_status" table
CREATE TABLE `new_volume_rustic_status` (
  `id` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `volume` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "volume_rustic_status" to new temporary table "new_volume_rustic_status"
INSERT INTO `new_volume_rustic_status` (`id`) SELECT `id` FROM `volume_rustic_status`;
-- drop "volume_rustic_status" table after copying rows
DROP TABLE `volume_rustic_status`;
-- rename temporary table "new_volume_rustic_status" to "volume_rustic_status"
ALTER TABLE `new_volume_rustic_status` RENAME TO `volume_rustic_status`;
-- create "new_box_volume_attachment" table
CREATE TABLE `new_box_volume_attachment` (
  `box_id` text NOT NULL,
  `volume_id` text NOT NULL,
  `root_uid` bigint NOT NULL,
  `root_gid` bigint NOT NULL,
  `root_mode` text NOT NULL,
  PRIMARY KEY (`box_id`, `volume_id`),
  CONSTRAINT `0` FOREIGN KEY (`volume_id`) REFERENCES `volume` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `1` FOREIGN KEY (`box_id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "box_volume_attachment" to new temporary table "new_box_volume_attachment"
INSERT INTO `new_box_volume_attachment` (`box_id`, `volume_id`, `root_uid`, `root_gid`, `root_mode`) SELECT `box_id`, `volume_id`, `root_uid`, `root_gid`, `root_mode` FROM `box_volume_attachment`;
-- drop "box_volume_attachment" table after copying rows
DROP TABLE `box_volume_attachment`;
-- rename temporary table "new_box_volume_attachment" to "box_volume_attachment"
ALTER TABLE `new_box_volume_attachment` RENAME TO `box_volume_attachment`;
-- create index "box_volume_attachment_volume_id" to table: "box_volume_attachment"
CREATE UNIQUE INDEX `box_volume_attachment_volume_id` ON `box_volume_attachment` (`volume_id`);
-- create "new_volume_snapshot_rustic" table
CREATE TABLE `new_volume_snapshot_rustic` (
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
-- copy rows from old table "volume_snapshot_rustic" to new temporary table "new_volume_snapshot_rustic"
INSERT INTO `new_volume_snapshot_rustic` (`id`, `snapshot_id`, `snapshot_time`, `parent_snapshot_id`, `hostname`, `files_new`, `files_changed`, `files_unmodified`, `total_files_processed`, `total_bytes_processed`, `dirs_new`, `dirs_changed`, `dirs_unmodified`, `total_dirs_processed`, `total_dirsize_processed`, `data_blobs`, `tree_blobs`, `data_added`, `data_added_packed`, `data_added_files`, `data_added_files_packed`, `data_added_trees`, `data_added_trees_packed`, `backup_start`, `backup_end`, `backup_duration`, `total_duration`) SELECT `id`, `snapshot_id`, `snapshot_time`, `parent_snapshot_id`, `hostname`, `files_new`, `files_changed`, `files_unmodified`, `total_files_processed`, `total_bytes_processed`, `dirs_new`, `dirs_changed`, `dirs_unmodified`, `total_dirs_processed`, `total_dirsize_processed`, `data_blobs`, `tree_blobs`, `data_added`, `data_added_packed`, `data_added_files`, `data_added_files_packed`, `data_added_trees`, `data_added_trees_packed`, `backup_start`, `backup_end`, `backup_duration`, `total_duration` FROM `volume_snapshot_rustic`;
-- drop "volume_snapshot_rustic" table after copying rows
DROP TABLE `volume_snapshot_rustic`;
-- rename temporary table "new_volume_snapshot_rustic" to "volume_snapshot_rustic"
ALTER TABLE `new_volume_snapshot_rustic` RENAME TO `volume_snapshot_rustic`;
-- create index "volume_snapshot_rustic_snapshot_id" to table: "volume_snapshot_rustic"
CREATE UNIQUE INDEX `volume_snapshot_rustic_snapshot_id` ON `volume_snapshot_rustic` (`snapshot_id`);
-- create "new_token" table
CREATE TABLE `new_token` (
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
-- copy rows from old table "token" to new temporary table "new_token"
INSERT INTO `new_token` (`id`, `workspace_id`, `created_at`, `name`, `token`, `for_workspace`, `box_id`) SELECT `id`, `workspace_id`, `created_at`, `name`, `token`, `for_workspace`, `box_id` FROM `token`;
-- drop "token" table after copying rows
DROP TABLE `token`;
-- rename temporary table "new_token" to "token"
ALTER TABLE `new_token` RENAME TO `token`;
-- create index "token_token" to table: "token"
CREATE UNIQUE INDEX `token_token` ON `token` (`token`);
-- create index "token_workspace_id_name" to table: "token"
CREATE UNIQUE INDEX `token_workspace_id_name` ON `token` (`workspace_id`, `name`);
-- create "new_log_metadata" table
CREATE TABLE `new_log_metadata` (
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
-- copy rows from old table "log_metadata" to new temporary table "new_log_metadata"
INSERT INTO `new_log_metadata` (`id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `box_id`, `file_name`, `format`, `metadata`, `total_line_bytes`, `last_log_time`) SELECT `id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `box_id`, `file_name`, `format`, `metadata`, `total_line_bytes`, `last_log_time` FROM `log_metadata`;
-- drop "log_metadata" table after copying rows
DROP TABLE `log_metadata`;
-- rename temporary table "new_log_metadata" to "log_metadata"
ALTER TABLE `new_log_metadata` RENAME TO `log_metadata`;
-- create index "log_metadata_box_id_file_name" to table: "log_metadata"
CREATE UNIQUE INDEX `log_metadata_box_id_file_name` ON `log_metadata` (`box_id`, `file_name`);
-- create "new_log_line" table
CREATE TABLE `new_log_line` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `workspace_id` text NOT NULL,
  `log_id` text NOT NULL,
  `time` datetime NOT NULL,
  `line` text NOT NULL,
  CONSTRAINT `0` FOREIGN KEY (`log_id`) REFERENCES `log_metadata` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "log_line" to new temporary table "new_log_line"
INSERT INTO `new_log_line` (`id`, `workspace_id`, `log_id`, `time`, `line`) SELECT `id`, `workspace_id`, `log_id`, `time`, `line` FROM `log_line`;
-- drop "log_line" table after copying rows
DROP TABLE `log_line`;
-- rename temporary table "new_log_line" to "log_line"
ALTER TABLE `new_log_line` RENAME TO `log_line`;
-- create index "log_line_log_id_and_id" to table: "log_line"
CREATE INDEX `log_line_log_id_and_id` ON `log_line` (`log_id`, `id`);
-- create index "log_line_time_index" to table: "log_line"
CREATE INDEX `log_line_time_index` ON `log_line` (`log_id`, `time`);
-- create "new_volume_snapshot" table
CREATE TABLE `new_volume_snapshot` (
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
-- copy rows from old table "volume_snapshot" to new temporary table "new_volume_snapshot"
INSERT INTO `new_volume_snapshot` (`id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `volume_provider_id`, `volume_id`, `lock_id`) SELECT `id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `volume_provider_id`, `volume_id`, `lock_id` FROM `volume_snapshot`;
-- drop "volume_snapshot" table after copying rows
DROP TABLE `volume_snapshot`;
-- rename temporary table "new_volume_snapshot" to "volume_snapshot"
ALTER TABLE `new_volume_snapshot` RENAME TO `volume_snapshot`;
-- create "new_box" table
CREATE TABLE `new_box` (
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
-- copy rows from old table "box" to new temporary table "new_box"
INSERT INTO `new_box` (`id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `name`, `network_id`, `network_type`, `dboxed_version`, `machine_id`, `desired_state`) SELECT `id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `name`, `network_id`, `network_type`, `dboxed_version`, `machine_id`, `desired_state` FROM `box`;
-- drop "box" table after copying rows
DROP TABLE `box`;
-- rename temporary table "new_box" to "box"
ALTER TABLE `new_box` RENAME TO `box`;
-- create index "box_workspace_id_name" to table: "box"
CREATE UNIQUE INDEX `box_workspace_id_name` ON `box` (`workspace_id`, `name`);
-- create "new_volume_provider_rustic" table
CREATE TABLE `new_volume_provider_rustic` (
  `id` text NOT NULL,
  `storage_type` text NOT NULL,
  `s3_bucket_id` text NULL,
  `storage_prefix` text NOT NULL,
  `password` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`s3_bucket_id`) REFERENCES `s3_bucket` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `1` FOREIGN KEY (`id`) REFERENCES `volume_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "volume_provider_rustic" to new temporary table "new_volume_provider_rustic"
INSERT INTO `new_volume_provider_rustic` (`id`, `storage_type`, `s3_bucket_id`, `storage_prefix`, `password`) SELECT `id`, `storage_type`, `s3_bucket_id`, `storage_prefix`, `password` FROM `volume_provider_rustic`;
-- drop "volume_provider_rustic" table after copying rows
DROP TABLE `volume_provider_rustic`;
-- rename temporary table "new_volume_provider_rustic" to "volume_provider_rustic"
ALTER TABLE `new_volume_provider_rustic` RENAME TO `volume_provider_rustic`;
-- create "new_s3_bucket" table
CREATE TABLE `new_s3_bucket` (
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
-- copy rows from old table "s3_bucket" to new temporary table "new_s3_bucket"
INSERT INTO `new_s3_bucket` (`id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `endpoint`, `bucket`, `access_key_id`, `secret_access_key`) SELECT `id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `endpoint`, `bucket`, `access_key_id`, `secret_access_key` FROM `s3_bucket`;
-- drop "s3_bucket" table after copying rows
DROP TABLE `s3_bucket`;
-- rename temporary table "new_s3_bucket" to "s3_bucket"
ALTER TABLE `new_s3_bucket` RENAME TO `s3_bucket`;
-- create index "s3_bucket_workspace_bucket" to table: "s3_bucket"
CREATE INDEX `s3_bucket_workspace_bucket` ON `s3_bucket` (`workspace_id`, `bucket`);
-- create "new_box_sandbox_status" table
CREATE TABLE `new_box_sandbox_status` (
  `id` text NOT NULL,
  `status_time` datetime NULL,
  `run_status` text NULL,
  `start_time` datetime NULL,
  `stop_time` datetime NULL,
  `docker_ps` blob NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "box_sandbox_status" to new temporary table "new_box_sandbox_status"
INSERT INTO `new_box_sandbox_status` (`id`, `status_time`, `run_status`, `start_time`, `stop_time`, `docker_ps`) SELECT `id`, `status_time`, `run_status`, `start_time`, `stop_time`, `docker_ps` FROM `box_sandbox_status`;
-- drop "box_sandbox_status" table after copying rows
DROP TABLE `box_sandbox_status`;
-- rename temporary table "new_box_sandbox_status" to "box_sandbox_status"
ALTER TABLE `new_box_sandbox_status` RENAME TO `box_sandbox_status`;
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
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create "volume_mount_status" table
DROP TABLE `volume_mount_status`;
-- reverse: create "new_box_sandbox_status" table
DROP TABLE `new_box_sandbox_status`;
-- reverse: create index "s3_bucket_workspace_bucket" to table: "s3_bucket"
DROP INDEX `s3_bucket_workspace_bucket`;
-- reverse: create "new_s3_bucket" table
DROP TABLE `new_s3_bucket`;
-- reverse: create "new_volume_provider_rustic" table
DROP TABLE `new_volume_provider_rustic`;
-- reverse: create index "box_workspace_id_name" to table: "box"
DROP INDEX `box_workspace_id_name`;
-- reverse: create "new_box" table
DROP TABLE `new_box`;
-- reverse: create "new_volume_snapshot" table
DROP TABLE `new_volume_snapshot`;
-- reverse: create index "log_line_time_index" to table: "log_line"
DROP INDEX `log_line_time_index`;
-- reverse: create index "log_line_log_id_and_id" to table: "log_line"
DROP INDEX `log_line_log_id_and_id`;
-- reverse: create "new_log_line" table
DROP TABLE `new_log_line`;
-- reverse: create index "log_metadata_box_id_file_name" to table: "log_metadata"
DROP INDEX `log_metadata_box_id_file_name`;
-- reverse: create "new_log_metadata" table
DROP TABLE `new_log_metadata`;
-- reverse: create index "token_workspace_id_name" to table: "token"
DROP INDEX `token_workspace_id_name`;
-- reverse: create index "token_token" to table: "token"
DROP INDEX `token_token`;
-- reverse: create "new_token" table
DROP TABLE `new_token`;
-- reverse: create index "volume_snapshot_rustic_snapshot_id" to table: "volume_snapshot_rustic"
DROP INDEX `volume_snapshot_rustic_snapshot_id`;
-- reverse: create "new_volume_snapshot_rustic" table
DROP TABLE `new_volume_snapshot_rustic`;
-- reverse: create index "box_volume_attachment_volume_id" to table: "box_volume_attachment"
DROP INDEX `box_volume_attachment_volume_id`;
-- reverse: create "new_box_volume_attachment" table
DROP TABLE `new_box_volume_attachment`;
-- reverse: create "new_volume_rustic_status" table
DROP TABLE `new_volume_rustic_status`;
-- reverse: create "new_volume_rustic" table
DROP TABLE `new_volume_rustic`;
-- reverse: create index "volume_workspace_id_name" to table: "volume"
DROP INDEX `volume_workspace_id_name`;
-- reverse: create "new_volume" table
DROP TABLE `new_volume`;
-- reverse: create index "volume_provider_workspace_id_name" to table: "volume_provider"
DROP INDEX `volume_provider_workspace_id_name`;
-- reverse: create "new_volume_provider" table
DROP TABLE `new_volume_provider`;
-- reverse: create "new_box_compose_project" table
DROP TABLE `new_box_compose_project`;
-- reverse: create "new_box_netbird" table
DROP TABLE `new_box_netbird`;
-- reverse: create "new_machine_hetzner_status" table
DROP TABLE `new_machine_hetzner_status`;
-- reverse: create "new_machine_hetzner" table
DROP TABLE `new_machine_hetzner`;
-- reverse: create index "machine_aws_status_instance_id" to table: "machine_aws_status"
DROP INDEX `machine_aws_status_instance_id`;
-- reverse: create "new_machine_aws_status" table
DROP TABLE `new_machine_aws_status`;
-- reverse: create "new_machine_aws" table
DROP TABLE `new_machine_aws`;
-- reverse: create index "machine_workspace_id_name" to table: "machine"
DROP INDEX `machine_workspace_id_name`;
-- reverse: create "new_machine" table
DROP TABLE `new_machine`;
-- reverse: create "new_network_netbird" table
DROP TABLE `new_network_netbird`;
-- reverse: create index "network_workspace_id_name" to table: "network"
DROP INDEX `network_workspace_id_name`;
-- reverse: create "new_network" table
DROP TABLE `new_network`;
-- reverse: create "new_machine_provider_hetzner_status" table
DROP TABLE `new_machine_provider_hetzner_status`;
-- reverse: create "new_machine_provider_hetzner" table
DROP TABLE `new_machine_provider_hetzner`;
-- reverse: create "new_machine_provider_aws_subnet" table
DROP TABLE `new_machine_provider_aws_subnet`;
-- reverse: create "new_machine_provider_aws_status" table
DROP TABLE `new_machine_provider_aws_status`;
-- reverse: create "new_machine_provider_aws" table
DROP TABLE `new_machine_provider_aws`;
-- reverse: create index "machine_provider_workspace_id_name" to table: "machine_provider"
DROP INDEX `machine_provider_workspace_id_name`;
-- reverse: create "new_machine_provider" table
DROP TABLE `new_machine_provider`;
-- reverse: create "new_workspace_quotas" table
DROP TABLE `new_workspace_quotas`;
-- reverse: create "new_workspace_access" table
DROP TABLE `new_workspace_access`;
-- reverse: create "new_workspace" table
DROP TABLE `new_workspace`;
-- reverse: create index "idx_table_and_id" to table: "change_tracking"
DROP INDEX `idx_table_and_id`;
-- reverse: create "new_change_tracking" table
DROP TABLE `new_change_tracking`;
