-- +goose Up
-- modify "box" table
ALTER TABLE "box" ALTER COLUMN "change_seq" DROP DEFAULT;
-- modify "box_netbird" table
ALTER TABLE "box_netbird" ALTER COLUMN "change_seq" DROP DEFAULT;
-- modify "load_balancer" table
ALTER TABLE "load_balancer" ALTER COLUMN "change_seq" DROP DEFAULT;
-- modify "machine" table
ALTER TABLE "machine" ALTER COLUMN "change_seq" DROP DEFAULT;
-- modify "machine_aws" table
ALTER TABLE "machine_aws" ALTER COLUMN "change_seq" DROP DEFAULT;
-- modify "machine_hetzner" table
ALTER TABLE "machine_hetzner" ALTER COLUMN "change_seq" DROP DEFAULT;
-- modify "machine_provider" table
ALTER TABLE "machine_provider" ALTER COLUMN "change_seq" DROP DEFAULT;
-- modify "network" table
ALTER TABLE "network" ALTER COLUMN "change_seq" DROP DEFAULT;
-- modify "s3_bucket" table
ALTER TABLE "s3_bucket" ALTER COLUMN "change_seq" DROP DEFAULT;
-- modify "volume" table
ALTER TABLE "volume" ALTER COLUMN "change_seq" DROP DEFAULT;
-- modify "volume_provider" table
ALTER TABLE "volume_provider" ALTER COLUMN "change_seq" DROP DEFAULT;
-- modify "workspace" table
ALTER TABLE "workspace" ALTER COLUMN "change_seq" DROP DEFAULT;

-- +goose Down
-- reverse: modify "workspace" table
ALTER TABLE "workspace" ALTER COLUMN "change_seq" SET DEFAULT 0;
-- reverse: modify "volume_provider" table
ALTER TABLE "volume_provider" ALTER COLUMN "change_seq" SET DEFAULT 0;
-- reverse: modify "volume" table
ALTER TABLE "volume" ALTER COLUMN "change_seq" SET DEFAULT 0;
-- reverse: modify "s3_bucket" table
ALTER TABLE "s3_bucket" ALTER COLUMN "change_seq" SET DEFAULT 0;
-- reverse: modify "network" table
ALTER TABLE "network" ALTER COLUMN "change_seq" SET DEFAULT 0;
-- reverse: modify "machine_provider" table
ALTER TABLE "machine_provider" ALTER COLUMN "change_seq" SET DEFAULT 0;
-- reverse: modify "machine_hetzner" table
ALTER TABLE "machine_hetzner" ALTER COLUMN "change_seq" SET DEFAULT 0;
-- reverse: modify "machine_aws" table
ALTER TABLE "machine_aws" ALTER COLUMN "change_seq" SET DEFAULT 0;
-- reverse: modify "machine" table
ALTER TABLE "machine" ALTER COLUMN "change_seq" SET DEFAULT 0;
-- reverse: modify "load_balancer" table
ALTER TABLE "load_balancer" ALTER COLUMN "change_seq" SET DEFAULT 0;
-- reverse: modify "box_netbird" table
ALTER TABLE "box_netbird" ALTER COLUMN "change_seq" SET DEFAULT 0;
-- reverse: modify "box" table
ALTER TABLE "box" ALTER COLUMN "change_seq" SET DEFAULT 0;
