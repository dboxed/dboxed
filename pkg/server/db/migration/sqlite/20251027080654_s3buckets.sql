-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_volume_provider_rustic" table
CREATE TABLE `new_volume_provider_rustic` (
  `id` bigint NOT NULL,
  `storage_type` text NOT NULL,
  `s3_bucket_id` bigint NULL,
  `storage_prefix` text NOT NULL,
  `password` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`s3_bucket_id`) REFERENCES `s3_bucket` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `1` FOREIGN KEY (`id`) REFERENCES `volume_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "volume_provider_rustic" to new temporary table "new_volume_provider_rustic"
INSERT INTO `new_volume_provider_rustic` (`id`, `storage_type`, `password`) SELECT `id`, `storage_type`, `password` FROM `volume_provider_rustic`;
-- drop "volume_provider_rustic" table after copying rows
DROP TABLE `volume_provider_rustic`;
-- rename temporary table "new_volume_provider_rustic" to "volume_provider_rustic"
ALTER TABLE `new_volume_provider_rustic` RENAME TO `volume_provider_rustic`;
-- drop "volume_provider_storage_s3" table
DROP TABLE `volume_provider_storage_s3`;
-- create "s3_bucket" table
CREATE TABLE `s3_bucket` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `workspace_id` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `reconcile_status` text NOT NULL DEFAULT 'Ok',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `endpoint` text NOT NULL,
  `bucket` text NOT NULL,
  `access_key_id` text NOT NULL,
  `secret_access_key` text NOT NULL,
  CONSTRAINT `0` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- create index "s3_bucket_workspace_bucket" to table: "s3_bucket"
CREATE INDEX `s3_bucket_workspace_bucket` ON `s3_bucket` (`workspace_id`, `bucket`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "s3_bucket_workspace_bucket" to table: "s3_bucket"
DROP INDEX `s3_bucket_workspace_bucket`;
-- reverse: create "s3_bucket" table
DROP TABLE `s3_bucket`;
-- reverse: drop "volume_provider_storage_s3" table
CREATE TABLE `volume_provider_storage_s3` (
  `id` bigint NOT NULL,
  `endpoint` text NOT NULL,
  `bucket` text NOT NULL,
  `access_key_id` text NOT NULL,
  `secret_access_key` text NOT NULL,
  `prefix` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `volume_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- reverse: create "new_volume_provider_rustic" table
DROP TABLE `new_volume_provider_rustic`;
