-- +goose Up
-- rename a column from "proxy_id" to "load_balancer_id"
ALTER TABLE `load_balancer_service` RENAME COLUMN `proxy_id` TO `load_balancer_id`;
-- rename a column from "proxy_type" to "load_balancer_type"
ALTER TABLE `load_balancer` RENAME COLUMN `proxy_type` TO `load_balancer_type`;
-- drop index "ingress_proxy_workspace_id_name" from table: "load_balancer"
DROP INDEX `ingress_proxy_workspace_id_name`;
-- create index "load_balancer_workspace_id_name" to table: "load_balancer"
CREATE UNIQUE INDEX `load_balancer_workspace_id_name` ON `load_balancer` (`workspace_id`, `name`);
-- rename a column from "ingress_proxy_id" to "load_balancer_id"
ALTER TABLE `load_balancer_box` RENAME COLUMN `ingress_proxy_id` TO `load_balancer_id`;

-- +goose Down
-- reverse: rename a column from "ingress_proxy_id" to "load_balancer_id"
ALTER TABLE `load_balancer_box` RENAME COLUMN `load_balancer_id` TO `ingress_proxy_id`;
-- reverse: create index "load_balancer_workspace_id_name" to table: "load_balancer"
DROP INDEX `load_balancer_workspace_id_name`;
-- reverse: drop index "ingress_proxy_workspace_id_name" from table: "load_balancer"
CREATE UNIQUE INDEX `ingress_proxy_workspace_id_name` ON `load_balancer` (`workspace_id`, `name`);
-- reverse: rename a column from "proxy_type" to "load_balancer_type"
ALTER TABLE `load_balancer` RENAME COLUMN `load_balancer_type` TO `proxy_type`;
-- reverse: rename a column from "proxy_id" to "load_balancer_id"
ALTER TABLE `load_balancer_service` RENAME COLUMN `load_balancer_id` TO `proxy_id`;
