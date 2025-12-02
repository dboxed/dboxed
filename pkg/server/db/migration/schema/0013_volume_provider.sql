create table volume_provider
(
    id                       text not null primary key,
    workspace_id             text           not null references workspace (id) on delete restrict,
    created_at               timestamptz not null default current_timestamp,
    deleted_at               timestamptz,
    finalizers               text           not null default '{}',

    reconcile_status         text           not null default 'Initializing',
    reconcile_status_details text           not null default '',

    type                     text           not null,
    name                     text           not null,

    unique (workspace_id, name)
);

create table volume_provider_rustic
(
    id             text not null primary key references volume_provider (id) on delete cascade,

    storage_type   text not null,
    s3_bucket_id   text references s3_bucket (id) on delete restrict,

    storage_prefix text not null,
    password       text not null
);
