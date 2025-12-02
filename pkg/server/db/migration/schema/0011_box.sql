create table box
(
    id                       text not null primary key,
    workspace_id             text           not null references workspace (id) on delete restrict,
    created_at               timestamptz not null default current_timestamp,
    deleted_at               timestamptz,
    finalizers               text           not null default '{}',

    reconcile_status         text           not null default 'Initializing',
    reconcile_status_details text           not null default '',

    name                     text           not null,
    box_type                 text           not null default 'normal',

    network_id               text references network (id) on delete restrict,
    network_type             text,
    dboxed_version           text           not null,

    machine_id               text           references machine (id) on delete set null,

    desired_state            text           not null default 'up',
    reconcile_requested_at   timestamptz,

    unique (workspace_id, name)
);

create table box_sandbox_status
(
    id          text not null primary key references box (id) on delete cascade,

    status_time timestamptz,

    run_status  text,
    start_time  timestamptz,
    stop_time   timestamptz,

    -- gzip compressed json
    docker_ps   bytea,

    network_ip4 text
);

alter table machine
    add foreign key (box_id) references box (id) on delete restrict;

create table box_netbird
(
    id                       text not null primary key references box (id) on delete cascade,

    reconcile_status         text not null default 'Initializing',
    reconcile_status_details text not null default '',

    setup_key_id             text,
    setup_key                text
);

create table box_compose_project
(
    box_id          text not null references box (id) on delete cascade,
    name            text not null,

    compose_project text not null,

    primary key (box_id, name)
);

create table box_port_forward
(
    id              text not null primary key,
    created_at      timestamptz not null default current_timestamp,

    box_id          text           not null references box (id) on delete cascade,
    description     text,

    protocol        text           not null,
    host_port_first int            not null,
    host_port_last  int            not null,
    sandbox_port    int            not null
);
