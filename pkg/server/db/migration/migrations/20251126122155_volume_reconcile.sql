-- +goose Up
-- modify "volume" table
ALTER TABLE "volume" ADD COLUMN "reconcile_status" text NOT NULL DEFAULT 'Initializing', ADD COLUMN "reconcile_status_details" text NOT NULL DEFAULT '';

-- +goose Down
-- reverse: modify "volume" table
ALTER TABLE "volume" DROP COLUMN "reconcile_status_details", DROP COLUMN "reconcile_status";
