-- +goose Up
-- add column "reconcile_status" to table: "volume_snapshot"
ALTER TABLE `volume_snapshot` ADD COLUMN `reconcile_status` text NOT NULL DEFAULT 'Initializing';
-- add column "reconcile_status_details" to table: "volume_snapshot"
ALTER TABLE `volume_snapshot` ADD COLUMN `reconcile_status_details` text NOT NULL DEFAULT '';

-- +goose Down
-- reverse: add column "reconcile_status_details" to table: "volume_snapshot"
ALTER TABLE `volume_snapshot` DROP COLUMN `reconcile_status_details`;
-- reverse: add column "reconcile_status" to table: "volume_snapshot"
ALTER TABLE `volume_snapshot` DROP COLUMN `reconcile_status`;
