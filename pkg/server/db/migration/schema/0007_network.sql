create table network
(
    id                       text        not null primary key,
    workspace_id             text        not null references workspace (id) on delete restrict,
    created_at               timestamptz not null default current_timestamp,
    deleted_at               timestamptz,
    finalizers               text        not null default '{}',

    reconcile_status         text        not null default 'Initializing',
    reconcile_status_details text        not null default '',

    type                     text        not null,
    name                     text        not null,

    unique (workspace_id, name)
);

create table network_netbird
(
    id               text not null primary key references network (id) on delete cascade,
    netbird_version  text not null,
    api_url          text not null,
    api_access_token text not null
);
