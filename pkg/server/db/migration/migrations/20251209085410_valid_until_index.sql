-- +goose Up
-- create index "token_valid_until" to table: "token"
CREATE INDEX "token_valid_until" ON "token" ("valid_until", "name");

-- +goose Down
-- reverse: create index "token_valid_until" to table: "token"
DROP INDEX "token_valid_until";
