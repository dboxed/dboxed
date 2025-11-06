create table machine_provider
(
    id                       TYPES_UUID_PRIMARY_KEY,
    workspace_id             text           not null references workspace (id) on delete restrict,
    created_at               TYPES_DATETIME not null default current_timestamp,
    deleted_at               TYPES_DATETIME,
    finalizers               text           not null default '{}',

    reconcile_status         text           not null default 'Initializing',
    reconcile_status_details text           not null default '',

    type                     text           not null,
    name                     text           not null,
    ssh_key_public           text,

    unique (workspace_id, name)
);
