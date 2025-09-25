-- +goose Up
-- modify "token" table
ALTER TABLE "token" DROP COLUMN "user_id", ADD COLUMN "workspace_id" bigint NOT NULL, ADD COLUMN "for_workspace" boolean NOT NULL, ADD COLUMN "box_id" bigint NULL, ADD CONSTRAINT "token_workspace_id_name_key" UNIQUE ("workspace_id", "name"), ADD CONSTRAINT "token_box_id_fkey" FOREIGN KEY ("box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "token_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
-- reverse: modify "token" table
ALTER TABLE "token" DROP CONSTRAINT "token_workspace_id_fkey", DROP CONSTRAINT "token_box_id_fkey", DROP CONSTRAINT "token_workspace_id_name_key", DROP COLUMN "box_id", DROP COLUMN "for_workspace", DROP COLUMN "workspace_id", ADD COLUMN "user_id" text NOT NULL;
