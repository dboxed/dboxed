-- +goose Up
-- modify "ingress_proxy" table
ALTER TABLE "ingress_proxy" DROP COLUMN "box_id", ADD COLUMN "network_id" text NOT NULL, ADD COLUMN "replicas" integer NOT NULL, ADD CONSTRAINT "ingress_proxy_network_id_fkey" FOREIGN KEY ("network_id") REFERENCES "network" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
-- create "ingress_proxy_box" table
CREATE TABLE "ingress_proxy_box" (
  "ingress_proxy_id" text NOT NULL,
  "box_id" text NOT NULL,
  CONSTRAINT "ingress_proxy_box_box_id_fkey" FOREIGN KEY ("box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "ingress_proxy_box_ingress_proxy_id_fkey" FOREIGN KEY ("ingress_proxy_id") REFERENCES "ingress_proxy" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- +goose Down
-- reverse: create "ingress_proxy_box" table
DROP TABLE "ingress_proxy_box";
-- reverse: modify "ingress_proxy" table
ALTER TABLE "ingress_proxy" DROP CONSTRAINT "ingress_proxy_network_id_fkey", DROP COLUMN "replicas", DROP COLUMN "network_id", ADD COLUMN "box_id" text NULL;
