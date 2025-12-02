-- +goose Up
-- modify "volume" table
ALTER TABLE "volume" DROP COLUMN "mount_time", DROP COLUMN "mount_box_id";
-- rename a column from "id" to "volume_id"
ALTER TABLE "volume_mount_status" RENAME COLUMN "id" TO "volume_id";
-- modify "volume_mount_status" table
ALTER TABLE "volume_mount_status" DROP CONSTRAINT "volume_mount_status_pkey", DROP CONSTRAINT "volume_mount_status_id_fkey", ALTER COLUMN "status_time" SET NOT NULL, DROP COLUMN "run_status", DROP COLUMN "start_time", DROP COLUMN "stop_time", DROP COLUMN "docker_ps", ADD COLUMN "mount_id" text NOT NULL, ADD COLUMN "box_id" text NULL, ADD COLUMN "mount_time" timestamptz NOT NULL, ADD COLUMN "release_time" timestamptz NULL, ADD COLUMN "force_released" boolean NOT NULL, ADD COLUMN "volume_total_size" bigint NULL, ADD COLUMN "volume_free_size" bigint NULL, ADD COLUMN "last_finished_snapshot_id" text NULL, ADD COLUMN "snapshot_start_time" timestamptz NULL, ADD COLUMN "snapshot_end_time" timestamptz NULL, ADD PRIMARY KEY ("volume_id", "mount_id"), ADD CONSTRAINT "volume_mount_status_mount_id_key" UNIQUE ("mount_id");
-- modify "volume" table
ALTER TABLE "volume" ADD CONSTRAINT "volume_id_mount_id_fkey" FOREIGN KEY ("id", "mount_id") REFERENCES "volume_mount_status" ("volume_id", "mount_id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- modify "volume_mount_status" table
ALTER TABLE "volume_mount_status" ADD CONSTRAINT "volume_mount_status_volume_id_fkey" FOREIGN KEY ("volume_id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
-- reverse: modify "volume_mount_status" table
ALTER TABLE "volume_mount_status" DROP CONSTRAINT "volume_mount_status_volume_id_fkey";
-- reverse: modify "volume" table
ALTER TABLE "volume" DROP CONSTRAINT "volume_id_mount_id_fkey";
-- reverse: modify "volume_mount_status" table
ALTER TABLE "volume_mount_status" DROP CONSTRAINT "volume_mount_status_mount_id_key", DROP CONSTRAINT "volume_mount_status_pkey", DROP COLUMN "snapshot_end_time", DROP COLUMN "snapshot_start_time", DROP COLUMN "last_finished_snapshot_id", DROP COLUMN "volume_free_size", DROP COLUMN "volume_total_size", DROP COLUMN "force_released", DROP COLUMN "release_time", DROP COLUMN "mount_time", DROP COLUMN "box_id", DROP COLUMN "mount_id", ADD COLUMN "docker_ps" bytea NULL, ADD COLUMN "stop_time" timestamptz NULL, ADD COLUMN "start_time" timestamptz NULL, ADD COLUMN "run_status" text NULL, ALTER COLUMN "status_time" DROP NOT NULL, ADD CONSTRAINT "volume_mount_status_id_fkey" FOREIGN KEY ("id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD PRIMARY KEY ("id");
-- reverse: rename a column from "id" to "volume_id"
ALTER TABLE "volume_mount_status" RENAME COLUMN "volume_id" TO "id";
-- reverse: modify "volume" table
ALTER TABLE "volume" ADD COLUMN "mount_box_id" text NULL, ADD COLUMN "mount_time" timestamptz NULL;
