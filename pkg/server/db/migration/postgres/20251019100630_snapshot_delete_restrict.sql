-- +goose Up
-- modify "volume_snapshot" table
ALTER TABLE "volume_snapshot" DROP CONSTRAINT "volume_snapshot_volume_id_fkey", ADD CONSTRAINT "volume_snapshot_volume_id_fkey" FOREIGN KEY ("volume_id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;

-- +goose Down
-- reverse: modify "volume_snapshot" table
ALTER TABLE "volume_snapshot" DROP CONSTRAINT "volume_snapshot_volume_id_fkey", ADD CONSTRAINT "volume_snapshot_volume_id_fkey" FOREIGN KEY ("volume_id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
