-- +goose Up
-- modify "box" table
ALTER TABLE "box" ADD COLUMN "desired_state" text NOT NULL DEFAULT 'stopped';

-- +goose Down
-- reverse: modify "box" table
ALTER TABLE "box" DROP COLUMN "desired_state";
