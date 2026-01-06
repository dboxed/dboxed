create table box_sandbox
(
    id           text        not null primary key,
    workspace_id text        not null references workspace (id) on delete restrict,
    box_id       text        not null references box (id) on delete cascade,
    created_at   timestamptz not null default current_timestamp,

    machine_id   text        not null,
    hostname     text        not null,

    status_time  timestamptz,

    run_status   text,
    start_time   timestamptz,
    stop_time    timestamptz,

    -- gzip compressed json
    docker_ps    bytea,

    network_ip4  text
);

alter table box
    add foreign key (current_sandbox_id) references box_sandbox (id) on delete set null;
