-- +goose Up
-- modify "box" table
ALTER TABLE "box" ADD COLUMN "box_type" text NOT NULL DEFAULT 'normal';
-- create "ingress_proxy" table
CREATE TABLE "ingress_proxy" (
  "id" text NOT NULL,
  "workspace_id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "finalizers" text NOT NULL DEFAULT '{}',
  "reconcile_status" text NOT NULL DEFAULT 'Initializing',
  "reconcile_status_details" text NOT NULL DEFAULT '',
  "box_id" text NULL,
  "name" text NOT NULL,
  "proxy_type" text NOT NULL,
  "http_port" integer NOT NULL,
  "https_port" integer NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "ingress_proxy_workspace_id_name_key" UNIQUE ("workspace_id", "name"),
  CONSTRAINT "ingress_proxy_box_id_fkey" FOREIGN KEY ("box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "ingress_proxy_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "box_ingress" table
CREATE TABLE "box_ingress" (
  "id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "proxy_id" text NOT NULL,
  "box_id" text NOT NULL,
  "description" text NULL,
  "hostname" text NOT NULL,
  "path_prefix" text NOT NULL,
  "port" integer NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "box_ingress_box_id_fkey" FOREIGN KEY ("box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "box_ingress_proxy_id_fkey" FOREIGN KEY ("proxy_id") REFERENCES "ingress_proxy" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT
);

-- +goose Down
-- reverse: create "box_ingress" table
DROP TABLE "box_ingress";
-- reverse: create "ingress_proxy" table
DROP TABLE "ingress_proxy";
-- reverse: modify "box" table
ALTER TABLE "box" DROP COLUMN "box_type";
