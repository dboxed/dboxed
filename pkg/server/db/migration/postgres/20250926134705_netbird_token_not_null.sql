-- +goose Up
-- modify "network_netbird" table
ALTER TABLE "network_netbird" ALTER COLUMN "api_access_token" SET NOT NULL;

-- +goose Down
-- reverse: modify "network_netbird" table
ALTER TABLE "network_netbird" ALTER COLUMN "api_access_token" DROP NOT NULL;
