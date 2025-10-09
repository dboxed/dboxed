-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- add column "total_line_bytes" to table: "log_metadata"
ALTER TABLE `log_metadata` ADD COLUMN `total_line_bytes` bigint NOT NULL DEFAULT 0;
-- create "new_log_line" table
CREATE TABLE `new_log_line` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `workspace_id` bigint NOT NULL,
  `log_id` bigint NOT NULL,
  `time` datetime NOT NULL,
  `line` text NOT NULL,
  CONSTRAINT `0` FOREIGN KEY (`log_id`) REFERENCES `log_metadata` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "log_line" to new temporary table "new_log_line"
INSERT INTO `new_log_line` (`id`, `log_id`, `time`, `line`) SELECT `id`, `log_id`, `time`, `line` FROM `log_line`;
-- drop "log_line" table after copying rows
DROP TABLE `log_line`;
-- rename temporary table "new_log_line" to "log_line"
ALTER TABLE `new_log_line` RENAME TO `log_line`;
-- create index "log_line_log_id_and_id" to table: "log_line"
CREATE INDEX `log_line_log_id_and_id` ON `log_line` (`log_id`, `id`);
-- create index "log_line_time_index" to table: "log_line"
CREATE INDEX `log_line_time_index` ON `log_line` (`log_id`, `time`);
-- create "workspace_quotas" table
CREATE TABLE `workspace_quotas` (
  `workspace_id` bigint NOT NULL,
  `max_log_bytes` int NOT NULL DEFAULT 100,
  PRIMARY KEY (`workspace_id`),
  CONSTRAINT `0` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create "workspace_quotas" table
DROP TABLE `workspace_quotas`;
-- reverse: create index "log_line_time_index" to table: "log_line"
DROP INDEX `log_line_time_index`;
-- reverse: create index "log_line_log_id_and_id" to table: "log_line"
DROP INDEX `log_line_log_id_and_id`;
-- reverse: create "new_log_line" table
DROP TABLE `new_log_line`;
-- reverse: add column "total_line_bytes" to table: "log_metadata"
ALTER TABLE `log_metadata` DROP COLUMN `total_line_bytes`;
