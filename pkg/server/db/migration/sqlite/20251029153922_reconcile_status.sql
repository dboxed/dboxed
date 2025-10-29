-- +goose Up
-- add column "reconcile_status" to table: "machine_aws"
ALTER TABLE `machine_aws` ADD COLUMN `reconcile_status` text NOT NULL DEFAULT 'Initializing';
-- add column "reconcile_status_details" to table: "machine_aws"
ALTER TABLE `machine_aws` ADD COLUMN `reconcile_status_details` text NOT NULL DEFAULT '';
-- add column "reconcile_status" to table: "machine_hetzner"
ALTER TABLE `machine_hetzner` ADD COLUMN `reconcile_status` text NOT NULL DEFAULT 'Initializing';
-- add column "reconcile_status_details" to table: "machine_hetzner"
ALTER TABLE `machine_hetzner` ADD COLUMN `reconcile_status_details` text NOT NULL DEFAULT '';
-- add column "reconcile_status" to table: "box_netbird"
ALTER TABLE `box_netbird` ADD COLUMN `reconcile_status` text NOT NULL DEFAULT 'Initializing';
-- add column "reconcile_status_details" to table: "box_netbird"
ALTER TABLE `box_netbird` ADD COLUMN `reconcile_status_details` text NOT NULL DEFAULT '';

-- +goose Down
-- reverse: add column "reconcile_status_details" to table: "box_netbird"
ALTER TABLE `box_netbird` DROP COLUMN `reconcile_status_details`;
-- reverse: add column "reconcile_status" to table: "box_netbird"
ALTER TABLE `box_netbird` DROP COLUMN `reconcile_status`;
-- reverse: add column "reconcile_status_details" to table: "machine_hetzner"
ALTER TABLE `machine_hetzner` DROP COLUMN `reconcile_status_details`;
-- reverse: add column "reconcile_status" to table: "machine_hetzner"
ALTER TABLE `machine_hetzner` DROP COLUMN `reconcile_status`;
-- reverse: add column "reconcile_status_details" to table: "machine_aws"
ALTER TABLE `machine_aws` DROP COLUMN `reconcile_status_details`;
-- reverse: add column "reconcile_status" to table: "machine_aws"
ALTER TABLE `machine_aws` DROP COLUMN `reconcile_status`;
