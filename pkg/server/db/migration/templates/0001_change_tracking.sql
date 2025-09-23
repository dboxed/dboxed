create table change_tracking
(
    id         TYPES_INT_PRIMARY_KEY,
    table_name text           not null,
    entity_id  bigint         not null,
    time       TYPES_DATETIME not null default current_timestamp
);

create index idx_table_and_id
    on change_tracking (table_name, id);
