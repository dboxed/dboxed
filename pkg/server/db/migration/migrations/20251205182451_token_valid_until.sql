-- +goose Up
-- modify "token" table
ALTER TABLE "token" ADD COLUMN "valid_until" timestamptz NULL;

-- +goose Down
-- reverse: modify "token" table
ALTER TABLE "token" DROP COLUMN "valid_until";
