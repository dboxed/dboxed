create table volume
(
    id                       text        not null primary key,
    workspace_id             text        not null references workspace (id) on delete restrict,
    created_at               timestamptz not null default current_timestamp,
    deleted_at               timestamptz,
    finalizers               text        not null default '{}',

    change_seq               bigint      not null,
    reconcile_status         text        not null default 'Initializing',
    reconcile_status_details text        not null default '',

    volume_provider_id       text        not null references volume_provider (id) on delete restrict,
    volume_provider_type     text        not null,

    name                     text        not null,

    mount_id                 text,

    -- we will later add the constraint
    latest_snapshot_id       text,

    unique (workspace_id, name)
);
create index volume_change_seq on volume (change_seq);

create table volume_restic
(
    id      text   not null primary key references volume (id) on delete cascade,
    fs_size bigint not null,
    fs_type text   not null
);

create table volume_restic_status
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
    volume_id                 text        not null references volume (id) on delete cascade,
    mount_id                  text        not null unique,

    -- we don't use a reference here so that deletion of boxes does not remove the box id
    box_id                    text,

    mount_time                timestamptz not null,
    release_time              timestamptz,
    force_released            bool        not null,

    status_time               timestamptz not null,

    volume_total_size         bigint,
    volume_free_size          bigint,

    -- we don't use a reference here so that deletion of snapshots does not remove the snapshot id
    last_finished_snapshot_id text,

    snapshot_start_time       timestamptz,
    snapshot_end_time         timestamptz,

    primary key (volume_id, mount_id)
);

alter table volume
    add foreign key (id, mount_id) references volume_mount_status (volume_id, mount_id) on delete restrict;
