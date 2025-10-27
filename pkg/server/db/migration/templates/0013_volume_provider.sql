create table volume_provider
(
    id                       TYPES_INT_PRIMARY_KEY,
    workspace_id             bigint         not null references workspace (id) on delete restrict,
    created_at               TYPES_DATETIME not null default current_timestamp,
    deleted_at               TYPES_DATETIME,
    finalizers               text           not null default '{}',

    reconcile_status         text           not null default 'Initializing',
    reconcile_status_details text           not null default '',

    type                     text           not null,
    name                     text           not null,

    unique (workspace_id, name)
);

create table volume_provider_rustic
(
    id           bigint not null primary key references volume_provider (id) on delete cascade,

    storage_type text   not null,
    s3_bucket_id bigint references s3_bucket (id) on delete restrict,

    storage_prefix text not null,
    password     text   not null
);
