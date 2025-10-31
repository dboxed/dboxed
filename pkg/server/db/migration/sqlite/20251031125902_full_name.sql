-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_user" table
CREATE TABLE `new_user` (
  `id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `username` text NULL,
  `email` text NULL,
  `full_name` text NOT NULL,
  `avatar` text NULL,
  PRIMARY KEY (`id`)
);
-- copy rows from old table "user" to new temporary table "new_user"
INSERT INTO `new_user` (`id`, `created_at`, `username`, `email`, `full_name`, `avatar`) SELECT `id`, `created_at`, `username`, `email`, `name`, `avatar` FROM `user`;
-- drop "user" table after copying rows
DROP TABLE `user`;
-- rename temporary table "new_user" to "user"
ALTER TABLE `new_user` RENAME TO `user`;
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create "new_user" table
DROP TABLE `new_user`;
