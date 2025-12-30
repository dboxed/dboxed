-- +goose Up
-- modify "box" table
ALTER TABLE "box" ADD COLUMN "current_sandbox_id" text NULL;
-- modify "box_sandbox_status" table
ALTER TABLE "box_sandbox_status" DROP CONSTRAINT "box_sandbox_status_id_fkey";
-- create "box_sandbox" table
CREATE TABLE "box_sandbox" (
  "id" text NOT NULL,
  "box_id" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "machine_id" text NOT NULL,
  "hostname" text NOT NULL,
  "status_time" timestamptz NULL,
  "run_status" text NULL,
  "start_time" timestamptz NULL,
  "stop_time" timestamptz NULL,
  "docker_ps" bytea NULL,
  "network_ip4" text NULL,
  PRIMARY KEY ("id")
);
-- modify "box" table
ALTER TABLE "box" ADD CONSTRAINT "box_current_sandbox_id_fkey" FOREIGN KEY ("current_sandbox_id") REFERENCES "box_sandbox" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "box_sandbox" table
ALTER TABLE "box_sandbox" ADD CONSTRAINT "box_sandbox_box_id_fkey" FOREIGN KEY ("box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- drop "box_sandbox_status" table
DROP TABLE "box_sandbox_status";

-- +goose Down
-- reverse: drop "box_sandbox_status" table
CREATE TABLE "box_sandbox_status" (
  "id" text NOT NULL,
  "status_time" timestamptz NULL,
  "run_status" text NULL,
  "start_time" timestamptz NULL,
  "stop_time" timestamptz NULL,
  "docker_ps" bytea NULL,
  "network_ip4" text NULL,
  PRIMARY KEY ("id")
);
-- reverse: modify "box_sandbox" table
ALTER TABLE "box_sandbox" DROP CONSTRAINT "box_sandbox_box_id_fkey";
-- reverse: modify "box" table
ALTER TABLE "box" DROP CONSTRAINT "box_current_sandbox_id_fkey";
-- reverse: create "box_sandbox" table
DROP TABLE "box_sandbox";
-- reverse: modify "box_sandbox_status" table
ALTER TABLE "box_sandbox_status" ADD CONSTRAINT "box_sandbox_status_id_fkey" FOREIGN KEY ("id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- reverse: modify "box" table
ALTER TABLE "box" DROP COLUMN "current_sandbox_id";
