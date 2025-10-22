-- +goose Up
-- modify "box" table
ALTER TABLE "box" ALTER COLUMN "desired_state" SET DEFAULT 'up';

-- +goose Down
-- reverse: modify "box" table
ALTER TABLE "box" ALTER COLUMN "desired_state" SET DEFAULT 'stopped';
