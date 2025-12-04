-- +goose Up
-- modify "log_metadata" table
ALTER TABLE "log_metadata" DROP CONSTRAINT "log_metadata_box_id_file_name_key", ADD COLUMN "machine_id" text NULL, ADD CONSTRAINT "log_metadata_machine_id_box_id_file_name_key" UNIQUE ("machine_id", "box_id", "file_name"), ADD CONSTRAINT "log_metadata_machine_id_fkey" FOREIGN KEY ("machine_id") REFERENCES "machine" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
-- reverse: modify "log_metadata" table
ALTER TABLE "log_metadata" DROP CONSTRAINT "log_metadata_machine_id_fkey", DROP CONSTRAINT "log_metadata_machine_id_box_id_file_name_key", DROP COLUMN "machine_id", ADD CONSTRAINT "log_metadata_box_id_file_name_key" UNIQUE ("box_id", "file_name");
