create table machine
(
    id                       text        not null primary key,
    workspace_id             text        not null references workspace (id) on delete restrict,
    created_at               timestamptz not null default current_timestamp,
    deleted_at               timestamptz,
    finalizers               text        not null default '{}',

    reconcile_status         text        not null default 'Initializing',
    reconcile_status_details text        not null default '',

    name                     text        not null,

    dboxed_version           text        not null,

    machine_provider_id      text references machine_provider (id) on delete restrict,
    machine_provider_type    text,

    unique (workspace_id, name)
);
