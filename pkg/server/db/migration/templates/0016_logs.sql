create table log_metadata
(
    id           TYPES_INT_PRIMARY_KEY,
    workspace_id bigint         not null references workspace (id) on delete cascade,
    created_at   TYPES_DATETIME not null default current_timestamp,
    deleted_at   TYPES_DATETIME,
    finalizers   text           not null default '{}',

    box_id       bigint references box (id) on delete cascade,

    file_name    text           not null,
    format       text           not null,
    metadata     text           not null,

    unique (box_id, file_name)
);

create table log_line
(
    id     TYPES_INT_PRIMARY_KEY,
    log_id bigint         not null references log_metadata (id) on delete cascade,

    time   TYPES_DATETIME not null,
    line   text           not null
);

create index log_line_log_id_and_id on log_line (log_id, id);
create index log_line_time_index on log_line (log_id, time);
