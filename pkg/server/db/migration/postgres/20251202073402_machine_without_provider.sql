-- +goose Up
-- modify "machine" table
ALTER TABLE "machine" ALTER COLUMN "machine_provider_id" DROP NOT NULL, ALTER COLUMN "machine_provider_type" DROP NOT NULL;

-- +goose Down
-- reverse: modify "machine" table
ALTER TABLE "machine" ALTER COLUMN "machine_provider_type" SET NOT NULL, ALTER COLUMN "machine_provider_id" SET NOT NULL;
