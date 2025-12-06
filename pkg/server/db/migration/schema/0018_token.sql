create table token
(
    id               text        not null primary key,
    workspace_id     text        not null references workspace (id) on delete cascade,
    created_at       timestamptz not null default current_timestamp,

    name             text        not null,
    type             text        not null,
    valid_until      timestamptz,
    token            text        not null unique,

    machine_id       text references machine (id) on delete cascade,
    box_id           text references box (id) on delete cascade,
    load_balancer_id text references load_balancer (id) on delete cascade,

    unique (workspace_id, name)
);
