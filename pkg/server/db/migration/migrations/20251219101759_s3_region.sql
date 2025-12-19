-- +goose Up
-- modify "s3_bucket" table
ALTER TABLE "s3_bucket" ADD COLUMN "determined_region" text NULL;

-- +goose Down
-- reverse: modify "s3_bucket" table
ALTER TABLE "s3_bucket" DROP COLUMN "determined_region";
