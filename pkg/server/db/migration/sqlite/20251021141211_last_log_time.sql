-- +goose Up
-- add column "last_log_time" to table: "log_metadata"
ALTER TABLE `log_metadata` ADD COLUMN `last_log_time` datetime NULL;

-- +goose Down
-- reverse: add column "last_log_time" to table: "log_metadata"
ALTER TABLE `log_metadata` DROP COLUMN `last_log_time`;
