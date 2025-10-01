-- +goose Up
-- modify "volume_snapshot" table
ALTER TABLE "volume_snapshot" ADD COLUMN "reconcile_status" text NOT NULL DEFAULT 'Initializing', ADD COLUMN "reconcile_status_details" text NOT NULL DEFAULT '';

-- +goose Down
-- reverse: modify "volume_snapshot" table
ALTER TABLE "volume_snapshot" DROP COLUMN "reconcile_status_details", DROP COLUMN "reconcile_status";
