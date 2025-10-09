-- +goose Up
-- modify "log_metadata" table
ALTER TABLE "log_metadata" ADD COLUMN "total_line_bytes" bigint NOT NULL DEFAULT 0;
-- modify "log_line" table
ALTER TABLE "log_line" ADD COLUMN "workspace_id" bigint NOT NULL, ADD CONSTRAINT "log_line_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- create "workspace_quotas" table
CREATE TABLE "workspace_quotas" (
  "workspace_id" bigint NOT NULL,
  "max_log_bytes" integer NOT NULL DEFAULT 100,
  PRIMARY KEY ("workspace_id"),
  CONSTRAINT "workspace_quotas_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- +goose Down
-- reverse: create "workspace_quotas" table
DROP TABLE "workspace_quotas";
-- reverse: modify "log_line" table
ALTER TABLE "log_line" DROP CONSTRAINT "log_line_workspace_id_fkey", DROP COLUMN "workspace_id";
-- reverse: modify "log_metadata" table
ALTER TABLE "log_metadata" DROP COLUMN "total_line_bytes";
