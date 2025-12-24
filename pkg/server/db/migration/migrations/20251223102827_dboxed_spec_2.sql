-- +goose Up
-- rename a constraint from "git_spec_pkey" to "dboxed_spec_pkey"
ALTER TABLE "dboxed_spec" RENAME CONSTRAINT "git_spec_pkey" TO "dboxed_spec_pkey";
-- modify "dboxed_spec" table
ALTER TABLE "dboxed_spec" DROP CONSTRAINT "git_spec_workspace_id_fkey", ADD CONSTRAINT "dboxed_spec_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
