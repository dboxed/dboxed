-- +goose Up
-- add column "network_ip4" to table: "box_sandbox_status"
ALTER TABLE `box_sandbox_status` ADD COLUMN `network_ip4` text NULL;

-- +goose Down
-- reverse: add column "network_ip4" to table: "box_sandbox_status"
ALTER TABLE `box_sandbox_status` DROP COLUMN `network_ip4`;
