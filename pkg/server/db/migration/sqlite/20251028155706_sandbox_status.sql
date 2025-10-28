-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- drop "box_run_status" table
DROP TABLE `box_run_status`;
-- create "box_sandbox_status" table
CREATE TABLE `box_sandbox_status` (
  `id` bigint NOT NULL,
  `status_time` datetime NULL,
  `run_status` text NULL,
  `start_time` datetime NULL,
  `stop_time` datetime NULL,
  `docker_ps` blob NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create "box_sandbox_status" table
DROP TABLE `box_sandbox_status`;
-- reverse: drop "box_run_status" table
CREATE TABLE `box_run_status` (
  `id` bigint NOT NULL,
  `status_time` datetime NULL,
  `run_status` text NULL,
  `start_time` datetime NULL,
  `stop_time` datetime NULL,
  `docker_ps` blob NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
