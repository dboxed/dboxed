-- +goose Up

drop table dboxed_spec_mapping;

-- create "dboxed_spec_mapping" table
CREATE TABLE "dboxed_spec_mapping" (
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "spec_id" text NOT NULL,
  "object_type" text NOT NULL,
  "object_id" text NOT NULL,
  "object_name" text NOT NULL,
  "recreate_key" text NOT NULL,
  "spec_fragment" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "dboxed_spec_mapping_workspace_id_spec_id_object_type_object_key" UNIQUE ("workspace_id", "spec_id", "object_type", "object_name"),
  CONSTRAINT "dboxed_spec_mapping_spec_id_fkey" FOREIGN KEY ("spec_id") REFERENCES "dboxed_spec" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "dboxed_spec_mapping_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- +goose Down
-- reverse: create "dboxed_spec_mapping" table
DROP TABLE "dboxed_spec_mapping";
