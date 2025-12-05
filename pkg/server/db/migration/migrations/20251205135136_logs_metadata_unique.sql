-- +goose Up
-- modify "log_metadata" table
ALTER TABLE "log_metadata" DROP CONSTRAINT "log_metadata_machine_id_box_id_file_name_key", ADD CONSTRAINT "log_metadata_machine_id_box_id_file_name_key" UNIQUE NULLS NOT DISTINCT ("machine_id", "box_id", "file_name");

-- +goose Down
-- reverse: modify "log_metadata" table
ALTER TABLE "log_metadata" DROP CONSTRAINT "log_metadata_machine_id_box_id_file_name_key", ADD CONSTRAINT "log_metadata_machine_id_box_id_file_name_key" UNIQUE ("machine_id", "box_id", "file_name");
