-- +goose Up
-- modify "log_metadata" table
ALTER TABLE "log_metadata" ADD COLUMN "sandbox_id" text NULL, ADD CONSTRAINT "log_metadata_sandbox_id_fkey" FOREIGN KEY ("sandbox_id") REFERENCES "box_sandbox" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
-- reverse: modify "log_metadata" table
ALTER TABLE "log_metadata" DROP CONSTRAINT "log_metadata_sandbox_id_fkey", DROP COLUMN "sandbox_id";
