-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
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
  `mount_id` text NULL,
  `latest_snapshot_id` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`, `mount_id`) REFERENCES `volume_mount_status` (`volume_id`, `mount_id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `1` FOREIGN KEY (`latest_snapshot_id`) REFERENCES `volume_snapshot` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `2` FOREIGN KEY (`volume_provider_id`) REFERENCES `volume_provider` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `3` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- copy rows from old table "volume" to new temporary table "new_volume"
INSERT INTO `new_volume` (`id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `volume_provider_id`, `volume_provider_type`, `name`, `mount_id`, `latest_snapshot_id`) SELECT `id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `volume_provider_id`, `volume_provider_type`, `name`, `mount_id`, `latest_snapshot_id` FROM `volume`;
-- drop "volume" table after copying rows
DROP TABLE `volume`;
-- rename temporary table "new_volume" to "volume"
ALTER TABLE `new_volume` RENAME TO `volume`;
-- create index "volume_workspace_id_name" to table: "volume"
CREATE UNIQUE INDEX `volume_workspace_id_name` ON `volume` (`workspace_id`, `name`);
-- create "new_volume_mount_status" table
CREATE TABLE `new_volume_mount_status` (
  `volume_id` text NOT NULL,
  `mount_id` text NOT NULL,
  `box_id` text NULL,
  `mount_time` datetime NOT NULL,
  `release_time` datetime NULL,
  `force_released` bool NOT NULL,
  `status_time` datetime NOT NULL,
  `volume_total_size` bigint NULL,
  `volume_free_size` bigint NULL,
  `last_finished_snapshot_id` text NULL,
  `snapshot_start_time` datetime NULL,
  `snapshot_end_time` datetime NULL,
  PRIMARY KEY (`volume_id`, `mount_id`),
  CONSTRAINT `0` FOREIGN KEY (`volume_id`) REFERENCES `volume` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "volume_mount_status" to new temporary table "new_volume_mount_status"
INSERT INTO `new_volume_mount_status` (`volume_id`, `status_time`) SELECT `id`, `status_time` FROM `volume_mount_status`;
-- drop "volume_mount_status" table after copying rows
DROP TABLE `volume_mount_status`;
-- rename temporary table "new_volume_mount_status" to "volume_mount_status"
ALTER TABLE `new_volume_mount_status` RENAME TO `volume_mount_status`;
-- create index "volume_mount_status_mount_id" to table: "volume_mount_status"
CREATE UNIQUE INDEX `volume_mount_status_mount_id` ON `volume_mount_status` (`mount_id`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "volume_mount_status_mount_id" to table: "volume_mount_status"
DROP INDEX `volume_mount_status_mount_id`;
-- reverse: create "new_volume_mount_status" table
DROP TABLE `new_volume_mount_status`;
-- reverse: create index "volume_workspace_id_name" to table: "volume"
DROP INDEX `volume_workspace_id_name`;
-- reverse: create "new_volume" table
DROP TABLE `new_volume`;
