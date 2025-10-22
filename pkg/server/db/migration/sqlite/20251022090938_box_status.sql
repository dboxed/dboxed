-- +goose Up
-- create "box_run_status" table
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

-- +goose Down
-- reverse: create "box_run_status" table
DROP TABLE `box_run_status`;
