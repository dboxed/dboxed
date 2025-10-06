-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_workspace" table
CREATE TABLE `new_workspace` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `name` text NOT NULL
);
-- copy rows from old table "workspace" to new temporary table "new_workspace"
INSERT INTO `new_workspace` (`id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `name`) SELECT `id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `name` FROM `workspace`;
-- drop "workspace" table after copying rows
DROP TABLE `workspace`;
-- rename temporary table "new_workspace" to "workspace"
ALTER TABLE `new_workspace` RENAME TO `workspace`;
-- create "new_box" table
CREATE TABLE `new_box` (
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
  `machine_id` bigint NULL,
  CONSTRAINT `0` FOREIGN KEY (`machine_id`) REFERENCES `machine` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `1` FOREIGN KEY (`network_id`) REFERENCES `network` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `2` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- copy rows from old table "box" to new temporary table "new_box"
INSERT INTO `new_box` (`id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `uuid`, `name`, `network_id`, `network_type`, `dboxed_version`, `box_spec`, `machine_id`) SELECT `id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `uuid`, `name`, `network_id`, `network_type`, `dboxed_version`, `box_spec`, `machine_id` FROM `box`;
-- drop "box" table after copying rows
DROP TABLE `box`;
-- rename temporary table "new_box" to "box"
ALTER TABLE `new_box` RENAME TO `box`;
-- create index "box_uuid" to table: "box"
CREATE UNIQUE INDEX `box_uuid` ON `box` (`uuid`);
-- create index "box_workspace_id_name" to table: "box"
CREATE UNIQUE INDEX `box_workspace_id_name` ON `box` (`workspace_id`, `name`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "box_workspace_id_name" to table: "box"
DROP INDEX `box_workspace_id_name`;
-- reverse: create index "box_uuid" to table: "box"
DROP INDEX `box_uuid`;
-- reverse: create "new_box" table
DROP TABLE `new_box`;
-- reverse: create "new_workspace" table
DROP TABLE `new_workspace`;
