-- +goose Up
-- modify "machine_aws_status" table
ALTER TABLE "machine_aws_status" ADD COLUMN "public_ip4" text NULL;
-- modify "machine_hetzner_status" table
ALTER TABLE "machine_hetzner_status" ADD COLUMN "public_ip4" text NULL;

-- +goose Down
-- reverse: modify "machine_hetzner_status" table
ALTER TABLE "machine_hetzner_status" DROP COLUMN "public_ip4";
-- reverse: modify "machine_aws_status" table
ALTER TABLE "machine_aws_status" DROP COLUMN "public_ip4";
