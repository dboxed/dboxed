-- +goose Up
-- modify "box_sandbox" table
ALTER TABLE "box_sandbox" ADD COLUMN "workspace_id" text NOT NULL, ADD CONSTRAINT "box_sandbox_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- modify "box" table
ALTER TABLE "box" DROP CONSTRAINT "box_current_sandbox_id_fkey", ADD CONSTRAINT "box_current_sandbox_id_fkey" FOREIGN KEY ("current_sandbox_id") REFERENCES "box_sandbox" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "box" table
ALTER TABLE "box" DROP CONSTRAINT "box_current_sandbox_id_fkey", ADD CONSTRAINT "box_current_sandbox_id_fkey" FOREIGN KEY ("current_sandbox_id") REFERENCES "box_sandbox" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- reverse: modify "box_sandbox" table
ALTER TABLE "box_sandbox" DROP CONSTRAINT "box_sandbox_workspace_id_fkey", DROP COLUMN "workspace_id";
