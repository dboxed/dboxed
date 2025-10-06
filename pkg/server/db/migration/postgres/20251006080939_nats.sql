-- +goose Up
-- modify "box" table
ALTER TABLE "box" DROP COLUMN "nkey", DROP COLUMN "nkey_seed";
-- modify "workspace" table
ALTER TABLE "workspace" DROP COLUMN "nkey", DROP COLUMN "nkey_seed";

-- +goose Down
-- reverse: modify "workspace" table
ALTER TABLE "workspace" ADD COLUMN "nkey_seed" text NOT NULL, ADD COLUMN "nkey" text NOT NULL;
-- reverse: modify "box" table
ALTER TABLE "box" ADD COLUMN "nkey_seed" text NOT NULL, ADD COLUMN "nkey" text NOT NULL;
