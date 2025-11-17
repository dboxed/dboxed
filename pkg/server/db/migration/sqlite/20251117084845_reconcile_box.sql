-- +goose Up
-- add column "reconcile_requested_at" to table: "box"
ALTER TABLE `box` ADD COLUMN `reconcile_requested_at` datetime NULL;

-- +goose Down
-- reverse: add column "reconcile_requested_at" to table: "box"
ALTER TABLE `box` DROP COLUMN `reconcile_requested_at`;
