-- +goose Up
-- modify "volume_provider_storage_s3" table
ALTER TABLE "volume_provider_storage_s3" DROP COLUMN "region";

-- +goose Down
-- reverse: modify "volume_provider_storage_s3" table
ALTER TABLE "volume_provider_storage_s3" ADD COLUMN "region" text NULL;
