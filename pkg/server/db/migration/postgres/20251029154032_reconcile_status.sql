-- +goose Up
-- modify "box_netbird" table
ALTER TABLE "box_netbird" ADD COLUMN "reconcile_status" text NOT NULL DEFAULT 'Initializing', ADD COLUMN "reconcile_status_details" text NOT NULL DEFAULT '';
-- modify "machine_aws" table
ALTER TABLE "machine_aws" ADD COLUMN "reconcile_status" text NOT NULL DEFAULT 'Initializing', ADD COLUMN "reconcile_status_details" text NOT NULL DEFAULT '';
-- modify "machine_hetzner" table
ALTER TABLE "machine_hetzner" ADD COLUMN "reconcile_status" text NOT NULL DEFAULT 'Initializing', ADD COLUMN "reconcile_status_details" text NOT NULL DEFAULT '';

-- +goose Down
-- reverse: modify "machine_hetzner" table
ALTER TABLE "machine_hetzner" DROP COLUMN "reconcile_status_details", DROP COLUMN "reconcile_status";
-- reverse: modify "machine_aws" table
ALTER TABLE "machine_aws" DROP COLUMN "reconcile_status_details", DROP COLUMN "reconcile_status";
-- reverse: modify "box_netbird" table
ALTER TABLE "box_netbird" DROP COLUMN "reconcile_status_details", DROP COLUMN "reconcile_status";
