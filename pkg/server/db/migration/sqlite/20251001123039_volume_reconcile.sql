-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_volume" table
CREATE TABLE `new_volume` (
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
-- copy rows from old table "volume" to new temporary table "new_volume"
INSERT INTO `new_volume` (`id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `volume_provider_id`, `volume_provider_type`, `uuid`, `name`, `lock_id`, `lock_time`, `lock_box_uuid`, `latest_snapshot_id`) SELECT `id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `volume_provider_id`, `volume_provider_type`, `uuid`, `name`, `lock_id`, `lock_time`, `lock_box_uuid`, `latest_snapshot_id` FROM `volume`;
-- drop "volume" table after copying rows
DROP TABLE `volume`;
-- rename temporary table "new_volume" to "volume"
ALTER TABLE `new_volume` RENAME TO `volume`;
-- create index "volume_uuid" to table: "volume"
CREATE UNIQUE INDEX `volume_uuid` ON `volume` (`uuid`);
-- create index "volume_workspace_id_name" to table: "volume"
CREATE UNIQUE INDEX `volume_workspace_id_name` ON `volume` (`workspace_id`, `name`);
-- create "new_volume_snapshot" table
CREATE TABLE `new_volume_snapshot` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `workspace_id` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `volume_provider_id` bigint NOT NULL,
  `volume_id` bigint NOT NULL,
  `lock_id` text NOT NULL,
  CONSTRAINT `0` FOREIGN KEY (`volume_id`) REFERENCES `volume` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`volume_provider_id`) REFERENCES `volume_provider` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `2` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- copy rows from old table "volume_snapshot" to new temporary table "new_volume_snapshot"
INSERT INTO `new_volume_snapshot` (`id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `volume_provider_id`, `volume_id`, `lock_id`) SELECT `id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `volume_provider_id`, `volume_id`, `lock_id` FROM `volume_snapshot`;
-- drop "volume_snapshot" table after copying rows
DROP TABLE `volume_snapshot`;
-- rename temporary table "new_volume_snapshot" to "volume_snapshot"
ALTER TABLE `new_volume_snapshot` RENAME TO `volume_snapshot`;
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create "new_volume_snapshot" table
DROP TABLE `new_volume_snapshot`;
-- reverse: create index "volume_workspace_id_name" to table: "volume"
DROP INDEX `volume_workspace_id_name`;
-- reverse: create index "volume_uuid" to table: "volume"
DROP INDEX `volume_uuid`;
-- reverse: create "new_volume" table
DROP TABLE `new_volume`;
