-- +goose Up
-- modify "volume" table
ALTER TABLE "volume" ADD COLUMN "lock_box_uuid" text NULL;

-- +goose Down
-- reverse: modify "volume" table
ALTER TABLE "volume" DROP COLUMN "lock_box_uuid";
