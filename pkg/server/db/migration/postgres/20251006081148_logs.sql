-- +goose Up
-- create "log_metadata" table
CREATE TABLE "log_metadata" (
  "id" bigserial NOT NULL,
  "workspace_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "finalizers" text NOT NULL DEFAULT '{}',
  "box_id" bigint NULL,
  "file_name" text NOT NULL,
  "format" text NOT NULL,
  "metadata" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "log_metadata_box_id_file_name_key" UNIQUE ("box_id", "file_name"),
  CONSTRAINT "log_metadata_box_id_fkey" FOREIGN KEY ("box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "log_metadata_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "log_line" table
CREATE TABLE "log_line" (
  "id" bigserial NOT NULL,
  "log_id" bigint NOT NULL,
  "time" timestamptz NOT NULL,
  "line" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "log_line_log_id_fkey" FOREIGN KEY ("log_id") REFERENCES "log_metadata" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- +goose Down
-- reverse: create "log_line" table
DROP TABLE "log_line";
-- reverse: create "log_metadata" table
DROP TABLE "log_metadata";
