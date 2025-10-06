-- +goose Up
-- create "log_metadata" table
CREATE TABLE `log_metadata` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `workspace_id` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `box_id` bigint NULL,
  `file_name` text NOT NULL,
  `format` text NOT NULL,
  `metadata` text NOT NULL,
  CONSTRAINT `0` FOREIGN KEY (`box_id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "log_metadata_box_id_file_name" to table: "log_metadata"
CREATE UNIQUE INDEX `log_metadata_box_id_file_name` ON `log_metadata` (`box_id`, `file_name`);
-- create "log_line" table
CREATE TABLE `log_line` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `log_id` bigint NOT NULL,
  `time` datetime NOT NULL,
  `line` text NOT NULL,
  CONSTRAINT `0` FOREIGN KEY (`log_id`) REFERENCES `log_metadata` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);

-- +goose Down
-- reverse: create "log_line" table
DROP TABLE `log_line`;
-- reverse: create index "log_metadata_box_id_file_name" to table: "log_metadata"
DROP INDEX `log_metadata_box_id_file_name`;
-- reverse: create "log_metadata" table
DROP TABLE `log_metadata`;
