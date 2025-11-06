create table volume
(
    id                   TYPES_UUID_PRIMARY_KEY,
    workspace_id         text           not null references workspace (id) on delete restrict,
    created_at           TYPES_DATETIME not null default current_timestamp,
    deleted_at           TYPES_DATETIME,
    finalizers           text           not null default '{}',

    volume_provider_id   text           not null references volume_provider (id) on delete restrict,
    volume_provider_type text           not null,

    name                 text           not null,

    lock_id              text,
    lock_time            TYPES_DATETIME,
    lock_box_id          text references box (id) on delete restrict,

    --{{ if eq .DbType "sqlite" }}
    latest_snapshot_id   text references volume_snapshot (id) on delete restrict,
    --{{ else }}
    -- we will later add the constraint
    latest_snapshot_id   text,
    --{{ end }}

    unique (workspace_id, name)
);

create table volume_rustic
(
    id      text   not null primary key references volume (id) on delete cascade,
    fs_size bigint not null,
    fs_type text   not null
);

create table volume_rustic_status
(
    id text not null primary key references volume (id) on delete cascade
);

create table box_volume_attachment
(
    box_id    text   not null references box (id) on delete cascade,
    volume_id text   not null references volume (id) on delete restrict unique,

    root_uid  bigint not null,
    root_gid  bigint not null,
    root_mode text   not null,

    primary key (box_id, volume_id)
);

create table volume_mount_status
(
    id          text not null primary key references volume (id) on delete cascade,


    status_time TYPES_DATETIME,

    run_status  text,
    start_time  TYPES_DATETIME,
    stop_time   TYPES_DATETIME,

    -- gzip compressed json
    docker_ps   TYPES_BYTES
);
