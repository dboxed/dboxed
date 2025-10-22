create table box
(
    id                       TYPES_INT_PRIMARY_KEY,
    workspace_id             bigint         not null references workspace (id) on delete restrict,
    created_at               TYPES_DATETIME not null default current_timestamp,
    deleted_at               TYPES_DATETIME,
    finalizers               text           not null default '{}',

    reconcile_status         text           not null default 'Initializing',
    reconcile_status_details text           not null default '',

    uuid                     text           not null default '' unique,
    name                     text           not null,

    network_id               bigint references network (id) on delete restrict,
    network_type             text,
    dboxed_version           text           not null,

    machine_id               bigint         references machine (id) on delete set null,

    unique (workspace_id, name)
);

create table box_run_status
(
    id          bigint not null primary key references box (id) on delete cascade,

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
    id           bigint not null primary key references box (id) on delete cascade,
    setup_key_id text,
    setup_key    text
);

create table box_compose_project
(
    box_id          bigint not null references box (id) on delete cascade,
    name            text   not null,

    compose_project text   not null,

    primary key (box_id, name)
);
