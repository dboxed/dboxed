-- +goose Up
-- create index "log_line_log_id_and_id" to table: "log_line"
CREATE INDEX `log_line_log_id_and_id` ON `log_line` (`log_id`, `id`);
-- create index "log_line_time_index" to table: "log_line"
CREATE INDEX `log_line_time_index` ON `log_line` (`log_id`, `time`);

-- +goose Down
-- reverse: create index "log_line_time_index" to table: "log_line"
DROP INDEX `log_line_time_index`;
-- reverse: create index "log_line_log_id_and_id" to table: "log_line"
DROP INDEX `log_line_log_id_and_id`;
