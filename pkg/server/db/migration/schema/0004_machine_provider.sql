create table machine_provider
(
    id                       text        not null primary key,
    workspace_id             text        not null references workspace (id) on delete restrict,
    created_at               timestamptz not null default current_timestamp,
    deleted_at               timestamptz,
    finalizers               text        not null default '{}',

    change_seq               bigint      not null,
    reconcile_status         text        not null default 'Initializing',
    reconcile_status_details text        not null default '',

    type                     text        not null,
    name                     text        not null,
    ssh_key_public           text,

    unique (workspace_id, name)
);
create index machine_provider_change_seq on machine_provider (change_seq);
