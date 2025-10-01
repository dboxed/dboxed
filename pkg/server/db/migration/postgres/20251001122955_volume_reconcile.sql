-- +goose Up
-- modify "volume" table
ALTER TABLE "volume" DROP COLUMN "reconcile_status", DROP COLUMN "reconcile_status_details";
-- modify "volume_snapshot" table
ALTER TABLE "volume_snapshot" DROP COLUMN "reconcile_status", DROP COLUMN "reconcile_status_details";

-- +goose Down
-- reverse: modify "volume_snapshot" table
ALTER TABLE "volume_snapshot" ADD COLUMN "reconcile_status_details" text NOT NULL DEFAULT '', ADD COLUMN "reconcile_status" text NOT NULL DEFAULT 'Initializing';
-- reverse: modify "volume" table
ALTER TABLE "volume" ADD COLUMN "reconcile_status_details" text NOT NULL DEFAULT '', ADD COLUMN "reconcile_status" text NOT NULL DEFAULT 'Initializing';
