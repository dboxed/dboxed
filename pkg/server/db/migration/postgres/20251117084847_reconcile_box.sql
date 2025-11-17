-- +goose Up
-- modify "box" table
ALTER TABLE "box" ADD COLUMN "reconcile_requested_at" timestamptz NULL;

-- +goose Down
-- reverse: modify "box" table
ALTER TABLE "box" DROP COLUMN "reconcile_requested_at";
