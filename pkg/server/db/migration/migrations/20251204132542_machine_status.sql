-- +goose Up
-- create "machine_run_status" table
CREATE TABLE "machine_run_status" (
  "id" text NOT NULL,
  "status_time" timestamptz NULL,
  "run_status" text NULL,
  "start_time" timestamptz NULL,
  "stop_time" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "machine_run_status_id_fkey" FOREIGN KEY ("id") REFERENCES "machine" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- +goose Down
-- reverse: create "machine_run_status" table
DROP TABLE "machine_run_status";
