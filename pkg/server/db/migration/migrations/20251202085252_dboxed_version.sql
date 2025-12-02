-- +goose Up
-- modify "box" table
ALTER TABLE "box" DROP COLUMN "dboxed_version";
-- modify "machine" table
ALTER TABLE "machine" ADD COLUMN "dboxed_version" text NOT NULL;

-- +goose Down
-- reverse: modify "machine" table
ALTER TABLE "machine" DROP COLUMN "dboxed_version";
-- reverse: modify "box" table
ALTER TABLE "box" ADD COLUMN "dboxed_version" text NOT NULL;
