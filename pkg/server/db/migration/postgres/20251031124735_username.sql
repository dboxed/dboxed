-- +goose Up
-- modify "user" table
ALTER TABLE "user" ADD COLUMN "username" text NOT NULL DEFAULT '';

-- +goose Down
-- reverse: modify "user" table
ALTER TABLE "user" DROP COLUMN "username";
