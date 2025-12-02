-- +goose Up
-- rename a column from "proxy_type" to "load_balancer_type"
ALTER TABLE "load_balancer" RENAME COLUMN "proxy_type" TO "load_balancer_type";
-- rename a constraint from "ingress_proxy_pkey" to "load_balancer_pkey"
ALTER TABLE "load_balancer" RENAME CONSTRAINT "ingress_proxy_pkey" TO "load_balancer_pkey";
-- modify "load_balancer" table
ALTER TABLE "load_balancer" DROP CONSTRAINT "ingress_proxy_network_id_fkey", DROP CONSTRAINT "ingress_proxy_workspace_id_fkey", ADD CONSTRAINT "load_balancer_network_id_fkey" FOREIGN KEY ("network_id") REFERENCES "network" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT, ADD CONSTRAINT "load_balancer_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- rename an index from "ingress_proxy_workspace_id_name_key" to "load_balancer_workspace_id_name_key"
ALTER INDEX "ingress_proxy_workspace_id_name_key" RENAME TO "load_balancer_workspace_id_name_key";
-- rename a column from "ingress_proxy_id" to "load_balancer_id"
ALTER TABLE "load_balancer_box" RENAME COLUMN "ingress_proxy_id" TO "load_balancer_id";
-- modify "load_balancer_box" table
ALTER TABLE "load_balancer_box" DROP CONSTRAINT "ingress_proxy_box_box_id_fkey", DROP CONSTRAINT "ingress_proxy_box_ingress_proxy_id_fkey", ADD CONSTRAINT "load_balancer_box_box_id_fkey" FOREIGN KEY ("box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "load_balancer_box_load_balancer_id_fkey" FOREIGN KEY ("load_balancer_id") REFERENCES "load_balancer" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- rename a column from "proxy_id" to "load_balancer_id"
ALTER TABLE "load_balancer_service" RENAME COLUMN "proxy_id" TO "load_balancer_id";
-- rename a constraint from "box_ingress_pkey" to "load_balancer_service_pkey"
ALTER TABLE "load_balancer_service" RENAME CONSTRAINT "box_ingress_pkey" TO "load_balancer_service_pkey";
-- modify "load_balancer_service" table
ALTER TABLE "load_balancer_service" DROP CONSTRAINT "box_ingress_box_id_fkey", DROP CONSTRAINT "box_ingress_proxy_id_fkey", ADD CONSTRAINT "load_balancer_service_box_id_fkey" FOREIGN KEY ("box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "load_balancer_service_load_balancer_id_fkey" FOREIGN KEY ("load_balancer_id") REFERENCES "load_balancer" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;

-- +goose Down
-- reverse: modify "load_balancer_service" table
ALTER TABLE "load_balancer_service" DROP CONSTRAINT "load_balancer_service_load_balancer_id_fkey", DROP CONSTRAINT "load_balancer_service_box_id_fkey", ADD CONSTRAINT "box_ingress_proxy_id_fkey" FOREIGN KEY ("proxy_id") REFERENCES "load_balancer" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT, ADD CONSTRAINT "box_ingress_box_id_fkey" FOREIGN KEY ("box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- reverse: rename a constraint from "box_ingress_pkey" to "load_balancer_service_pkey"
ALTER TABLE "load_balancer_service" RENAME CONSTRAINT "load_balancer_service_pkey" TO "box_ingress_pkey";
-- reverse: rename a column from "proxy_id" to "load_balancer_id"
ALTER TABLE "load_balancer_service" RENAME COLUMN "load_balancer_id" TO "proxy_id";
-- reverse: modify "load_balancer_box" table
ALTER TABLE "load_balancer_box" DROP CONSTRAINT "load_balancer_box_load_balancer_id_fkey", DROP CONSTRAINT "load_balancer_box_box_id_fkey", ADD CONSTRAINT "ingress_proxy_box_ingress_proxy_id_fkey" FOREIGN KEY ("ingress_proxy_id") REFERENCES "load_balancer" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "ingress_proxy_box_box_id_fkey" FOREIGN KEY ("box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- reverse: rename a column from "ingress_proxy_id" to "load_balancer_id"
ALTER TABLE "load_balancer_box" RENAME COLUMN "load_balancer_id" TO "ingress_proxy_id";
-- reverse: rename an index from "ingress_proxy_workspace_id_name_key" to "load_balancer_workspace_id_name_key"
ALTER INDEX "load_balancer_workspace_id_name_key" RENAME TO "ingress_proxy_workspace_id_name_key";
-- reverse: modify "load_balancer" table
ALTER TABLE "load_balancer" DROP CONSTRAINT "load_balancer_workspace_id_fkey", DROP CONSTRAINT "load_balancer_network_id_fkey", ADD CONSTRAINT "ingress_proxy_workspace_id_fkey" FOREIGN KEY ("workspace_id") REFERENCES "workspace" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "ingress_proxy_network_id_fkey" FOREIGN KEY ("network_id") REFERENCES "network" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- reverse: rename a constraint from "ingress_proxy_pkey" to "load_balancer_pkey"
ALTER TABLE "load_balancer" RENAME CONSTRAINT "load_balancer_pkey" TO "ingress_proxy_pkey";
-- reverse: rename a column from "proxy_type" to "load_balancer_type"
ALTER TABLE "load_balancer" RENAME COLUMN "load_balancer_type" TO "proxy_type";
