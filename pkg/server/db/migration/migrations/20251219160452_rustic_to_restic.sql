-- +goose Up
alter table "volume_provider_rustic" rename to "volume_provider_restic";
alter table "volume_rustic" rename to "volume_restic";
alter table "volume_rustic_status" rename to "volume_restic_status";
alter table "volume_snapshot_rustic" rename to "volume_snapshot_restic";

-- rename a constraint from "volume_provider_rustic_pkey" to "volume_provider_restic_pkey"
ALTER TABLE "volume_provider_restic" RENAME CONSTRAINT "volume_provider_rustic_pkey" TO "volume_provider_restic_pkey";
-- modify "volume_provider_restic" table
ALTER TABLE "volume_provider_restic" DROP CONSTRAINT "volume_provider_rustic_id_fkey", DROP CONSTRAINT "volume_provider_rustic_s3_bucket_id_fkey", ADD CONSTRAINT "volume_provider_restic_id_fkey" FOREIGN KEY ("id") REFERENCES "volume_provider" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "volume_provider_restic_s3_bucket_id_fkey" FOREIGN KEY ("s3_bucket_id") REFERENCES "s3_bucket" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- rename a constraint from "volume_rustic_pkey" to "volume_restic_pkey"
ALTER TABLE "volume_restic" RENAME CONSTRAINT "volume_rustic_pkey" TO "volume_restic_pkey";
-- modify "volume_restic" table
ALTER TABLE "volume_restic" DROP CONSTRAINT "volume_rustic_id_fkey", ADD CONSTRAINT "volume_restic_id_fkey" FOREIGN KEY ("id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- rename a constraint from "volume_rustic_status_pkey" to "volume_restic_status_pkey"
ALTER TABLE "volume_restic_status" RENAME CONSTRAINT "volume_rustic_status_pkey" TO "volume_restic_status_pkey";
-- modify "volume_restic_status" table
ALTER TABLE "volume_restic_status" DROP CONSTRAINT "volume_rustic_status_id_fkey", ADD CONSTRAINT "volume_restic_status_id_fkey" FOREIGN KEY ("id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- rename a constraint from "volume_snapshot_rustic_pkey" to "volume_snapshot_restic_pkey"
ALTER TABLE "volume_snapshot_restic" RENAME CONSTRAINT "volume_snapshot_rustic_pkey" TO "volume_snapshot_restic_pkey";
-- modify "volume_snapshot_restic" table
ALTER TABLE "volume_snapshot_restic" DROP CONSTRAINT "volume_snapshot_rustic_id_fkey", ADD CONSTRAINT "volume_snapshot_restic_id_fkey" FOREIGN KEY ("id") REFERENCES "volume_snapshot" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- rename an index from "volume_snapshot_rustic_snapshot_id_key" to "volume_snapshot_restic_snapshot_id_key"
ALTER INDEX "volume_snapshot_rustic_snapshot_id_key" RENAME TO "volume_snapshot_restic_snapshot_id_key";

-- +goose Down
-- reverse: rename an index from "volume_snapshot_rustic_snapshot_id_key" to "volume_snapshot_restic_snapshot_id_key"
ALTER INDEX "volume_snapshot_restic_snapshot_id_key" RENAME TO "volume_snapshot_rustic_snapshot_id_key";
-- reverse: modify "volume_snapshot_restic" table
ALTER TABLE "volume_snapshot_restic" DROP CONSTRAINT "volume_snapshot_restic_id_fkey", ADD CONSTRAINT "volume_snapshot_rustic_id_fkey" FOREIGN KEY ("id") REFERENCES "volume_snapshot" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- reverse: rename a constraint from "volume_snapshot_rustic_pkey" to "volume_snapshot_restic_pkey"
ALTER TABLE "volume_snapshot_restic" RENAME CONSTRAINT "volume_snapshot_restic_pkey" TO "volume_snapshot_rustic_pkey";
-- reverse: modify "volume_restic_status" table
ALTER TABLE "volume_restic_status" DROP CONSTRAINT "volume_restic_status_id_fkey", ADD CONSTRAINT "volume_rustic_status_id_fkey" FOREIGN KEY ("id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- reverse: rename a constraint from "volume_rustic_status_pkey" to "volume_restic_status_pkey"
ALTER TABLE "volume_restic_status" RENAME CONSTRAINT "volume_restic_status_pkey" TO "volume_rustic_status_pkey";
-- reverse: modify "volume_restic" table
ALTER TABLE "volume_restic" DROP CONSTRAINT "volume_restic_id_fkey", ADD CONSTRAINT "volume_rustic_id_fkey" FOREIGN KEY ("id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- reverse: rename a constraint from "volume_rustic_pkey" to "volume_restic_pkey"
ALTER TABLE "volume_restic" RENAME CONSTRAINT "volume_restic_pkey" TO "volume_rustic_pkey";
-- reverse: modify "volume_provider_restic" table
ALTER TABLE "volume_provider_restic" DROP CONSTRAINT "volume_provider_restic_s3_bucket_id_fkey", DROP CONSTRAINT "volume_provider_restic_id_fkey", ADD CONSTRAINT "volume_provider_rustic_s3_bucket_id_fkey" FOREIGN KEY ("s3_bucket_id") REFERENCES "s3_bucket" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT, ADD CONSTRAINT "volume_provider_rustic_id_fkey" FOREIGN KEY ("id") REFERENCES "volume_provider" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- reverse: rename a constraint from "volume_provider_rustic_pkey" to "volume_provider_restic_pkey"
ALTER TABLE "volume_provider_restic" RENAME CONSTRAINT "volume_provider_restic_pkey" TO "volume_provider_rustic_pkey";

alter table "volume_provider_restic" rename to "volume_provider_rustic";
alter table "volume_restic" rename to "volume_rustic";
alter table "volume_restic_status" rename to "volume_rustic_status";
alter table "volume_snapshot_restic" rename to "volume_snapshot_rustic";
