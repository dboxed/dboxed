-- +goose Up
-- create "box_sandbox_status" table
CREATE TABLE "box_sandbox_status" (
  "id" bigint NOT NULL,
  "status_time" timestamptz NULL,
  "run_status" text NULL,
  "start_time" timestamptz NULL,
  "stop_time" timestamptz NULL,
  "docker_ps" bytea NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "box_sandbox_status_id_fkey" FOREIGN KEY ("id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- drop "box_run_status" table
DROP TABLE "box_run_status";

-- +goose Down
-- reverse: drop "box_run_status" table
CREATE TABLE "box_run_status" (
  "id" bigint NOT NULL,
  "status_time" timestamptz NULL,
  "run_status" text NULL,
  "start_time" timestamptz NULL,
  "stop_time" timestamptz NULL,
  "docker_ps" bytea NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "box_run_status_id_fkey" FOREIGN KEY ("id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- reverse: create "box_sandbox_status" table
DROP TABLE "box_sandbox_status";
