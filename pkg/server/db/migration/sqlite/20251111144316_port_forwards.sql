-- +goose Up
-- create "box_port_forward" table
CREATE TABLE `box_port_forward` (
  `id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `box_id` text NOT NULL,
  `description` text NULL,
  `protocol` text NOT NULL,
  `host_port_first` int NOT NULL,
  `host_port_last` int NOT NULL,
  `sandbox_port` int NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`box_id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);

-- +goose Down
-- reverse: create "box_port_forward" table
DROP TABLE `box_port_forward`;
