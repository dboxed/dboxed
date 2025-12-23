-- +goose Up
alter table git_spec rename to dboxed_spec;
alter table git_spec_mapping rename to dboxed_spec_mapping;
update box set box_type = 'normal' where box_type = 'git-spec';

-- +goose Down
alter table dboxed_spec rename to git_spec;
alter table dboxed_spec_mapping rename to git_spec_mapping;
