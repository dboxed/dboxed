create table machine
(
    id                       TYPES_UUID_PRIMARY_KEY,
    workspace_id             text           not null references workspace (id) on delete restrict,
    created_at               TYPES_DATETIME not null default current_timestamp,
    deleted_at               TYPES_DATETIME,
    finalizers               text           not null default '{}',

    reconcile_status         text           not null default 'Initializing',
    reconcile_status_details text           not null default '',

    name                     text           not null,
    machine_provider_id      text           not null references machine_provider (id) on delete restrict,
    machine_provider_type    text           not null,

    --{{ if eq .DbType "sqlite" }}
    box_id                   text           not null references box (id) on delete restrict,
    --{{ else }}
    -- we will later add the constraint
    box_id                   text           not null,
    --{{ end }}

    unique (workspace_id, name)
);
