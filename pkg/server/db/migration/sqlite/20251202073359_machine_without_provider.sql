-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
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
  `machine_provider_id` text NULL,
  `machine_provider_type` text NULL,
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
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "machine_workspace_id_name" to table: "machine"
DROP INDEX `machine_workspace_id_name`;
-- reverse: create "new_machine" table
DROP TABLE `new_machine`;
