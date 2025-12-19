-- +goose Up
update volume_provider set type = 'restic' where type = 'rustic';
update volume set volume_provider_type = 'restic' where volume_provider_type = 'rustic';
