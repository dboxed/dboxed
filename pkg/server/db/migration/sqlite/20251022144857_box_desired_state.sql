-- +goose Up
-- add column "desired_state" to table: "box"
ALTER TABLE `box` ADD COLUMN `desired_state` text NOT NULL DEFAULT 'stopped';

-- +goose Down
-- reverse: add column "desired_state" to table: "box"
ALTER TABLE `box` DROP COLUMN `desired_state`;
