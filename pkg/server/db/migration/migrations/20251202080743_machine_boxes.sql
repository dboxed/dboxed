-- +goose Up
-- modify "machine" table
ALTER TABLE "machine" DROP COLUMN "box_id";

-- +goose Down
-- reverse: modify "machine" table
ALTER TABLE "machine" ADD COLUMN "box_id" text NOT NULL;
