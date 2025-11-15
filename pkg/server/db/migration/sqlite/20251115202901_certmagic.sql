-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_token" table
CREATE TABLE `new_token` (
  `id` text NOT NULL,
  `workspace_id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `name` text NOT NULL,
  `token` text NOT NULL,
  `for_workspace` bool NOT NULL,
  `box_id` text NULL,
  `load_balancer_id` text NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`load_balancer_id`) REFERENCES `load_balancer` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`box_id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `2` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "token" to new temporary table "new_token"
INSERT INTO `new_token` (`id`, `workspace_id`, `created_at`, `name`, `token`, `for_workspace`, `box_id`) SELECT `id`, `workspace_id`, `created_at`, `name`, `token`, `for_workspace`, `box_id` FROM `token`;
-- drop "token" table after copying rows
DROP TABLE `token`;
-- rename temporary table "new_token" to "token"
ALTER TABLE `new_token` RENAME TO `token`;
-- create index "token_token" to table: "token"
CREATE UNIQUE INDEX `token_token` ON `token` (`token`);
-- create index "token_workspace_id_name" to table: "token"
CREATE UNIQUE INDEX `token_workspace_id_name` ON `token` (`workspace_id`, `name`);
-- create "load_balancer_certmagic" table
CREATE TABLE `load_balancer_certmagic` (
  `load_balancer_id` text NOT NULL,
  `key` text NOT NULL,
  `value` blob NOT NULL,
  `last_modified` datetime NOT NULL,
  CONSTRAINT `0` FOREIGN KEY (`load_balancer_id`) REFERENCES `load_balancer` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "load_balancer_certmagic_load_balancer_id_key" to table: "load_balancer_certmagic"
CREATE UNIQUE INDEX `load_balancer_certmagic_load_balancer_id_key` ON `load_balancer_certmagic` (`load_balancer_id`, `key`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "load_balancer_certmagic_load_balancer_id_key" to table: "load_balancer_certmagic"
DROP INDEX `load_balancer_certmagic_load_balancer_id_key`;
-- reverse: create "load_balancer_certmagic" table
DROP TABLE `load_balancer_certmagic`;
-- reverse: create index "token_workspace_id_name" to table: "token"
DROP INDEX `token_workspace_id_name`;
-- reverse: create index "token_token" to table: "token"
DROP INDEX `token_token`;
-- reverse: create "new_token" table
DROP TABLE `new_token`;
