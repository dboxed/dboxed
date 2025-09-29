create table network
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

create table network_netbird
(
    id               bigint not null primary key references network (id) on delete cascade,
    netbird_version  text not null,
    api_url          text not null,
    api_access_token text not null
);
