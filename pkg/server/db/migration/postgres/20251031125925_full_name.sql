-- +goose Up
-- rename a column from "name" to "full_name"
ALTER TABLE "user" RENAME COLUMN "name" TO "full_name";
-- modify "user" table
ALTER TABLE "user" ALTER COLUMN "username" DROP NOT NULL, ALTER COLUMN "username" DROP DEFAULT;

-- +goose Down
-- reverse: modify "user" table
ALTER TABLE "user" ALTER COLUMN "username" SET NOT NULL, ALTER COLUMN "username" SET DEFAULT '';
-- reverse: rename a column from "name" to "full_name"
ALTER TABLE "user" RENAME COLUMN "full_name" TO "name";
