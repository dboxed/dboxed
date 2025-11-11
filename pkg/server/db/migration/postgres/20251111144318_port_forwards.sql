-- +goose Up
-- create "box_port_forward" table
CREATE TABLE "box_port_forward" (
  "id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "box_id" text NOT NULL,
  "description" text NULL,
  "protocol" text NOT NULL,
  "host_port_first" integer NOT NULL,
  "host_port_last" integer NOT NULL,
  "sandbox_port" integer NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "box_port_forward_box_id_fkey" FOREIGN KEY ("box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- +goose Down
-- reverse: create "box_port_forward" table
DROP TABLE "box_port_forward";
