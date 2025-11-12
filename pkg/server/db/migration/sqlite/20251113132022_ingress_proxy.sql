-- +goose Up
-- add column "box_type" to table: "box"
ALTER TABLE `box` ADD COLUMN `box_type` text NOT NULL DEFAULT 'normal';
-- create "ingress_proxy" table
CREATE TABLE `ingress_proxy` (
  `id` text NOT NULL,
  `workspace_id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `box_id` text NULL,
  `name` text NOT NULL,
  `proxy_type` text NOT NULL,
  `http_port` int NOT NULL,
  `https_port` int NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`box_id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `1` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "ingress_proxy_workspace_id_name" to table: "ingress_proxy"
CREATE UNIQUE INDEX `ingress_proxy_workspace_id_name` ON `ingress_proxy` (`workspace_id`, `name`);
-- create "box_ingress" table
CREATE TABLE `box_ingress` (
  `id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `proxy_id` text NOT NULL,
  `box_id` text NOT NULL,
  `description` text NULL,
  `hostname` text NOT NULL,
  `path_prefix` text NOT NULL,
  `port` int NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`box_id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`proxy_id`) REFERENCES `ingress_proxy` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT
);

-- +goose Down
-- reverse: create "box_ingress" table
DROP TABLE `box_ingress`;
-- reverse: create index "ingress_proxy_workspace_id_name" to table: "ingress_proxy"
DROP INDEX `ingress_proxy_workspace_id_name`;
-- reverse: create "ingress_proxy" table
DROP TABLE `ingress_proxy`;
-- reverse: add column "box_type" to table: "box"
ALTER TABLE `box` DROP COLUMN `box_type`;
