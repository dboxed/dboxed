-- +goose Up
-- modify "volume_snapshot" table
ALTER TABLE "volume_snapshot" DROP CONSTRAINT "volume_snapshot_volume_id_fkey", ALTER COLUMN "volume_id" DROP NOT NULL, ADD CONSTRAINT "volume_snapshot_volume_id_fkey" FOREIGN KEY ("volume_id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "volume_snapshot" table
ALTER TABLE "volume_snapshot" DROP CONSTRAINT "volume_snapshot_volume_id_fkey", ALTER COLUMN "volume_id" SET NOT NULL, ADD CONSTRAINT "volume_snapshot_volume_id_fkey" FOREIGN KEY ("volume_id") REFERENCES "volume" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
