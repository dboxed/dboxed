-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_ingress_proxy" table
CREATE TABLE `new_ingress_proxy` (
  `id` text NOT NULL,
  `workspace_id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp),
  `deleted_at` datetime NULL,
  `finalizers` text NOT NULL DEFAULT '{}',
  `reconcile_status` text NOT NULL DEFAULT 'Initializing',
  `reconcile_status_details` text NOT NULL DEFAULT '',
  `name` text NOT NULL,
  `proxy_type` text NOT NULL,
  `network_id` text NOT NULL,
  `replicas` int NOT NULL,
  `http_port` int NOT NULL,
  `https_port` int NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`network_id`) REFERENCES `network` (`id`) ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT `1` FOREIGN KEY (`workspace_id`) REFERENCES `workspace` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "ingress_proxy" to new temporary table "new_ingress_proxy"
INSERT INTO `new_ingress_proxy` (`id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `name`, `proxy_type`, `http_port`, `https_port`) SELECT `id`, `workspace_id`, `created_at`, `deleted_at`, `finalizers`, `reconcile_status`, `reconcile_status_details`, `name`, `proxy_type`, `http_port`, `https_port` FROM `ingress_proxy`;
-- drop "ingress_proxy" table after copying rows
DROP TABLE `ingress_proxy`;
-- rename temporary table "new_ingress_proxy" to "ingress_proxy"
ALTER TABLE `new_ingress_proxy` RENAME TO `ingress_proxy`;
-- create index "ingress_proxy_workspace_id_name" to table: "ingress_proxy"
CREATE UNIQUE INDEX `ingress_proxy_workspace_id_name` ON `ingress_proxy` (`workspace_id`, `name`);
-- create "ingress_proxy_box" table
CREATE TABLE `ingress_proxy_box` (
  `ingress_proxy_id` text NOT NULL,
  `box_id` text NOT NULL,
  CONSTRAINT `0` FOREIGN KEY (`box_id`) REFERENCES `box` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`ingress_proxy_id`) REFERENCES `ingress_proxy` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create "ingress_proxy_box" table
DROP TABLE `ingress_proxy_box`;
-- reverse: create index "ingress_proxy_workspace_id_name" to table: "ingress_proxy"
DROP INDEX `ingress_proxy_workspace_id_name`;
-- reverse: create "new_ingress_proxy" table
DROP TABLE `new_ingress_proxy`;
