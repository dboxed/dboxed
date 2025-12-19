-- +goose Up
-- modify "volume_snapshot_restic" table
ALTER TABLE "volume_snapshot_restic" DROP COLUMN "total_dirs_processed", DROP COLUMN "total_dirsize_processed", DROP COLUMN "data_added_files", DROP COLUMN "data_added_files_packed", DROP COLUMN "data_added_trees", DROP COLUMN "data_added_trees_packed", DROP COLUMN "backup_duration", DROP COLUMN "total_duration";

-- +goose Down
-- reverse: modify "volume_snapshot_restic" table
ALTER TABLE "volume_snapshot_restic" ADD COLUMN "total_duration" real NOT NULL, ADD COLUMN "backup_duration" real NOT NULL, ADD COLUMN "data_added_trees_packed" integer NOT NULL, ADD COLUMN "data_added_trees" integer NOT NULL, ADD COLUMN "data_added_files_packed" integer NOT NULL, ADD COLUMN "data_added_files" integer NOT NULL, ADD COLUMN "total_dirsize_processed" integer NOT NULL, ADD COLUMN "total_dirs_processed" integer NOT NULL;
