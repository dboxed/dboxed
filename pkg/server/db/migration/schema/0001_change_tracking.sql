create table change_tracking
(
    id         bigserial   not null primary key,
    table_name text        not null,
    entity_id  text        not null,
    time       timestamptz not null default current_timestamp
);

create index idx_table_and_id
    on change_tracking (table_name, id);
