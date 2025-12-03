-- +goose Up
-- modify "token" table
ALTER TABLE "token" ADD COLUMN "machine_id" text NULL, ADD CONSTRAINT "token_machine_id_fkey" FOREIGN KEY ("machine_id") REFERENCES "machine" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
-- reverse: modify "token" table
ALTER TABLE "token" DROP CONSTRAINT "token_machine_id_fkey", DROP COLUMN "machine_id";
