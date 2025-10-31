-- +goose Up
-- add column "username" to table: "user"
ALTER TABLE `user` ADD COLUMN `username` text NOT NULL DEFAULT '';

-- +goose Down
-- reverse: add column "username" to table: "user"
ALTER TABLE `user` DROP COLUMN `username`;
