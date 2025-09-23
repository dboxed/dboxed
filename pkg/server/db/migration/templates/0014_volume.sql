create table volume
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

    volume_provider_id       bigint         not null references volume_provider (id) on delete restrict,
    volume_provider_type     text           not null,

    unique (workspace_id, name)
);

create table volume_dboxed
(
    id      bigint primary key references volume (id) on delete cascade,

    fs_size bigint not null,
    fs_type text   not null
);

create table volume_dboxed_status
(
    id        bigint primary key references volume (id) on delete cascade,

    volume_id bigint,
    fs_size   bigint,
    fs_type   text
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
