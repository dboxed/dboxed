-- +goose Up
create sequence change_tracking_seq;

-- +goose Down
drop sequence change_tracking_seq;
