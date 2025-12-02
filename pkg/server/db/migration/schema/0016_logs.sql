create table log_metadata
(
    id               text        not null primary key,
    workspace_id     text        not null references workspace (id) on delete cascade,
    created_at       timestamptz not null default current_timestamp,
    deleted_at       timestamptz,
    finalizers       text        not null default '{}',

    box_id           text references box (id) on delete cascade,

    file_name        text        not null,
    format           text        not null,
    metadata         text        not null,

    total_line_bytes bigint      not null default 0,
    last_log_time    timestamptz,

    unique (box_id, file_name)
);

create table log_line
(
    id           bigserial   not null primary key,
    workspace_id text        not null references workspace (id) on delete cascade,

    log_id       text        not null references log_metadata (id) on delete cascade,

    time         timestamptz not null,
    line         text        not null
);

create index log_line_log_id_and_id on log_line (log_id, id);
create index log_line_time_index on log_line (log_id, time);
