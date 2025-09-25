-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_token" table
CREATE TABLE `new_token` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `workspace_id` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `name` text NOT NULL,
  `token` text NOT NULL,
  `for_workspace` bool NOT NULL,
  `box_id` bigint NULL,
  CONSTRAINT `0` FOREIGN KEY (`box_id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "token" to new temporary table "new_token"
INSERT INTO `new_token` (`id`, `created_at`, `name`, `token`) SELECT `id`, `created_at`, `name`, `token` FROM `token`;
-- drop "token" table after copying rows
DROP TABLE `token`;
-- rename temporary table "new_token" to "token"
ALTER TABLE `new_token` RENAME TO `token`;
-- create index "token_token" to table: "token"
CREATE UNIQUE INDEX `token_token` ON `token` (`token`);
-- create index "token_workspace_id_name" to table: "token"
CREATE UNIQUE INDEX `token_workspace_id_name` ON `token` (`workspace_id`, `name`);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create index "token_workspace_id_name" to table: "token"
DROP INDEX `token_workspace_id_name`;
-- reverse: create index "token_token" to table: "token"
DROP INDEX `token_token`;
-- reverse: create "new_token" table
DROP TABLE `new_token`;
