-- +goose Up
-- create "load_balancer_certmagic" table
CREATE TABLE "load_balancer_certmagic" (
  "load_balancer_id" text NOT NULL,
  "key" text NOT NULL,
  "value" bytea NOT NULL,
  "last_modified" timestamptz NOT NULL,
  CONSTRAINT "load_balancer_certmagic_load_balancer_id_key_key" UNIQUE ("load_balancer_id", "key"),
  CONSTRAINT "load_balancer_certmagic_load_balancer_id_fkey" FOREIGN KEY ("load_balancer_id") REFERENCES "load_balancer" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- modify "token" table
ALTER TABLE "token" ADD COLUMN "load_balancer_id" text NULL, ADD CONSTRAINT "token_load_balancer_id_fkey" FOREIGN KEY ("load_balancer_id") REFERENCES "load_balancer" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
-- reverse: modify "token" table
ALTER TABLE "token" DROP CONSTRAINT "token_load_balancer_id_fkey", DROP COLUMN "load_balancer_id";
-- reverse: create "load_balancer_certmagic" table
DROP TABLE "load_balancer_certmagic";
