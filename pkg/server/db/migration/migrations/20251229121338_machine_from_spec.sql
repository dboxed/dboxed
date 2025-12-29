-- +goose Up
-- modify "box" table
ALTER TABLE "box" ADD COLUMN "machine_from_spec" boolean NOT NULL DEFAULT false;

-- +goose Down
-- reverse: modify "box" table
ALTER TABLE "box" DROP COLUMN "machine_from_spec";
