create table token
(
    id            TYPES_INT_PRIMARY_KEY,
    workspace_id  bigint         not null references workspace (id) on delete cascade,
    created_at    TYPES_DATETIME not null default current_timestamp,

    name          text           not null,
    token         text           not null unique,

    for_workspace bool           not null,
    box_id        bigint references box (id) on delete cascade,

    unique (workspace_id, name)
);
