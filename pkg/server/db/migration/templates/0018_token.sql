create table token
(
    id               TYPES_UUID_PRIMARY_KEY,
    workspace_id     text           not null references workspace (id) on delete cascade,
    created_at       TYPES_DATETIME not null default current_timestamp,

    name             text           not null,
    token            text           not null unique,

    for_workspace    bool           not null,
    box_id           text references box (id) on delete cascade,
    load_balancer_id text references load_balancer (id) on delete cascade,

    unique (workspace_id, name)
);
