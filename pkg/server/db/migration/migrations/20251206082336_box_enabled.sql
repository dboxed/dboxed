-- +goose Up
-- modify "box" table
ALTER TABLE "box" DROP COLUMN "desired_state", ADD COLUMN "enabled" boolean NOT NULL DEFAULT true;

-- +goose Down
-- reverse: modify "box" table
ALTER TABLE "box" DROP COLUMN "enabled", ADD COLUMN "desired_state" text NOT NULL DEFAULT 'up';
