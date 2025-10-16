-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_volume_provider_storage_s3" table
CREATE TABLE `new_volume_provider_storage_s3` (
  `id` bigint NOT NULL,
  `endpoint` text NOT NULL,
  `bucket` text NOT NULL,
  `access_key_id` text NOT NULL,
  `secret_access_key` text NOT NULL,
  `prefix` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `volume_provider` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "volume_provider_storage_s3" to new temporary table "new_volume_provider_storage_s3"
INSERT INTO `new_volume_provider_storage_s3` (`id`, `endpoint`, `bucket`, `access_key_id`, `secret_access_key`, `prefix`) SELECT `id`, `endpoint`, `bucket`, `access_key_id`, `secret_access_key`, `prefix` FROM `volume_provider_storage_s3`;
-- drop "volume_provider_storage_s3" table after copying rows
DROP TABLE `volume_provider_storage_s3`;
-- rename temporary table "new_volume_provider_storage_s3" to "volume_provider_storage_s3"
ALTER TABLE `new_volume_provider_storage_s3` RENAME TO `volume_provider_storage_s3`;
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create "new_volume_provider_storage_s3" table
DROP TABLE `new_volume_provider_storage_s3`;
