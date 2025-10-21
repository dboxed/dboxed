-- +goose Up
-- modify "log_metadata" table
ALTER TABLE "log_metadata" ADD COLUMN "last_log_time" timestamptz NULL;

-- +goose Down
-- reverse: modify "log_metadata" table
ALTER TABLE "log_metadata" DROP COLUMN "last_log_time";
