create table box
(
    id                       TYPES_UUID_PRIMARY_KEY,
    workspace_id             text           not null references workspace (id) on delete restrict,
    created_at               TYPES_DATETIME not null default current_timestamp,
    deleted_at               TYPES_DATETIME,
    finalizers               text           not null default '{}',

    reconcile_status         text           not null default 'Initializing',
    reconcile_status_details text           not null default '',

    name                     text           not null,

    network_id               text references network (id) on delete restrict,
    network_type             text,
    dboxed_version           text           not null,

    machine_id               text           references machine (id) on delete set null,

    desired_state            text           not null default 'up',

    unique (workspace_id, name)
);

create table box_sandbox_status
(
    id          text not null primary key references box (id) on delete cascade,

    status_time TYPES_DATETIME,

    run_status  text,
    start_time  TYPES_DATETIME,
    stop_time   TYPES_DATETIME,

    -- gzip compressed json
    docker_ps   TYPES_BYTES
);

--{{ if eq .DbType "postgres" }}
alter table machine
    add foreign key (box_id) references box (id) on delete restrict;
--{{ end }}

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
