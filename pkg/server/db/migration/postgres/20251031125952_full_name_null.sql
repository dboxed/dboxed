-- +goose Up
-- modify "user" table
ALTER TABLE "user" ALTER COLUMN "full_name" DROP NOT NULL;

-- +goose Down
-- reverse: modify "user" table
ALTER TABLE "user" ALTER COLUMN "full_name" SET NOT NULL;
