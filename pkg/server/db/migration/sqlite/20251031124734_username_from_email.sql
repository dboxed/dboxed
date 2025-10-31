-- +goose Up
-- add column "username" to table: "user"
UPDATE `user` SET username = email;
