-- +goose Up
-- modify "box_sandbox_status" table
ALTER TABLE "box_sandbox_status" ADD COLUMN "network_ip4" text NULL;

-- +goose Down
-- reverse: modify "box_sandbox_status" table
ALTER TABLE "box_sandbox_status" DROP COLUMN "network_ip4";
