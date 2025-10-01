create table volume
(
    id                       TYPES_INT_PRIMARY_KEY,
    workspace_id             bigint         not null references workspace (id) on delete restrict,
    created_at               TYPES_DATETIME not null default current_timestamp,
    deleted_at               TYPES_DATETIME,
    finalizers               text           not null default '{}',

    volume_provider_id       bigint         not null references volume_provider (id) on delete restrict,
    volume_provider_type     text           not null,

    uuid                     text           not null unique,
    name                     text           not null,

    lock_id                  text,
    lock_time                TYPES_DATETIME,
    lock_box_uuid            text,

    --{{ if eq .DbType "sqlite" }}
    latest_snapshot_id       bigint references volume_snapshot (id) on delete restrict,
    --{{ else }}
    -- we will later add the constraint
    latest_snapshot_id       bigint,
    --{{ end }}

    unique (workspace_id, name)
);

create table volume_rustic
(
    id      bigint not null primary key references volume (id) on delete cascade,
    fs_size bigint not null,
    fs_type text   not null
);

create table volume_rustic_status
(
    id bigint not null primary key references volume (id) on delete cascade
);

create table box_volume_attachment
(
    box_id    bigint not null references box (id) on delete cascade,
    volume_id bigint not null references volume (id) on delete restrict unique,

    root_uid  bigint not null,
    root_gid  bigint not null,
    root_mode text   not null,

    primary key (box_id, volume_id)
);
