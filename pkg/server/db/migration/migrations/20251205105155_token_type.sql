-- +goose Up
-- modify "token" table
delete from "token";
ALTER TABLE "token" DROP COLUMN "for_workspace", ADD COLUMN "type" text NOT NULL;

-- +goose Down
-- reverse: modify "token" table
ALTER TABLE "token" DROP COLUMN "type", ADD COLUMN "for_workspace" boolean NOT NULL;
