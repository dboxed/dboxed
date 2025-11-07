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

    mount_id             text,

    --{{ if eq .DbType "sqlite" }}
    latest_snapshot_id   text references volume_snapshot (id) on delete restrict,
    --{{ else }}
    -- we will later add the constraint
    latest_snapshot_id   text,
    --{{ end }}

    unique (workspace_id, name)

    --{{ if eq .DbType "sqlite" }}
    ,
    foreign key (id, mount_id) references volume_mount_status (volume_id, mount_id) on delete restrict
    --{{ end }}
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
    volume_id                 text           not null references volume (id) on delete cascade,
    mount_id                  text           not null unique,

    -- we don't use a reference here so that deletion of boxes does not remove the box id
    box_id                    text,

    mount_time                TYPES_DATETIME not null,
    release_time              TYPES_DATETIME,
    force_released            bool           not null,

    status_time               TYPES_DATETIME not null,

    volume_total_size         bigint,
    volume_free_size          bigint,

    -- we don't use a reference here so that deletion of snapshots does not remove the snapshot id
    last_finished_snapshot_id text,

    snapshot_start_time       TYPES_DATETIME,
    snapshot_end_time         TYPES_DATETIME,

    primary key (volume_id, mount_id)
);

--{{ if eq .DbType "postgres" }}
alter table volume
    add foreign key (id, mount_id) references volume_mount_status (volume_id, mount_id) on delete restrict;
--{{ end }}
