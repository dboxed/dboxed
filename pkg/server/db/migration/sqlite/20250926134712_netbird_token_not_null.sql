-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "new_network_netbird" table
CREATE TABLE `new_network_netbird` (
  `id` bigint NOT NULL,
  `netbird_version` text NOT NULL,
  `api_url` text NOT NULL,
  `api_access_token` text NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`id`) REFERENCES `network` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- copy rows from old table "network_netbird" to new temporary table "new_network_netbird"
INSERT INTO `new_network_netbird` (`id`, `netbird_version`, `api_url`, `api_access_token`) SELECT `id`, `netbird_version`, `api_url`, `api_access_token` FROM `network_netbird`;
-- drop "network_netbird" table after copying rows
DROP TABLE `network_netbird`;
-- rename temporary table "new_network_netbird" to "network_netbird"
ALTER TABLE `new_network_netbird` RENAME TO `network_netbird`;
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create "new_network_netbird" table
DROP TABLE `new_network_netbird`;
