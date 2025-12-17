-- +goose Up
-- create "git_credentials" table
CREATE TABLE "git_credentials" (
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "host" text NOT NULL,
  "path_glob" text NOT NULL,
  "credentials_type" text NOT NULL,
  "username" text NULL,
  "password" text NULL,
  "ssh_key" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "git_credentials_workspace_id_host_path_glob_key" UNIQUE ("workspace_id", "host", "path_glob"),
  CONSTRAINT "git_credentials_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "git_spec" table
CREATE TABLE "git_spec" (
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "finalizers" text NOT NULL DEFAULT '{}',
  "change_seq" bigint NOT NULL,
  "reconcile_status" text NOT NULL DEFAULT 'Initializing',
  "reconcile_status_details" text NOT NULL DEFAULT '',
  "git_url" text NOT NULL,
  "git_ref" text NULL,
  "subdir" text NOT NULL,
  "spec_file" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "git_spec_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "git_spec_mapping" table
CREATE TABLE "git_spec_mapping" (
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "repo_key" text NOT NULL,
  "recreate_key" text NOT NULL,
  "object_type" text NOT NULL,
  "object_id" text NOT NULL,
  "object_name" text NOT NULL,
  "spec" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "git_spec_mapping_workspace_id_repo_key_object_type_object_n_key" UNIQUE ("workspace_id", "repo_key", "object_type", "object_name"),
  CONSTRAINT "git_spec_mapping_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- +goose Down
-- reverse: create "git_spec_mapping" table
DROP TABLE "git_spec_mapping";
-- reverse: create "git_spec" table
DROP TABLE "git_spec";
-- reverse: create "git_credentials" table
DROP TABLE "git_credentials";
