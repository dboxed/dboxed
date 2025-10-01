-- +goose Up
-- add column "lock_box_uuid" to table: "volume"
ALTER TABLE `volume` ADD COLUMN `lock_box_uuid` text NULL;

-- +goose Down
-- reverse: add column "lock_box_uuid" to table: "volume"
ALTER TABLE `volume` DROP COLUMN `lock_box_uuid`;
