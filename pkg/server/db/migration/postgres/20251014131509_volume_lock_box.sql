-- +goose Up
-- modify "volume" table
ALTER TABLE "volume" DROP COLUMN "lock_box_uuid", ADD COLUMN "lock_box_id" bigint NULL, ADD CONSTRAINT "volume_lock_box_id_fkey" FOREIGN KEY ("lock_box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;

-- +goose Down
-- reverse: modify "volume" table
ALTER TABLE "volume" DROP CONSTRAINT "volume_lock_box_id_fkey", DROP COLUMN "lock_box_id", ADD COLUMN "lock_box_uuid" text NULL;
