create table workspace
(
    id                       text        not null primary key,
    created_at               timestamptz not null default current_timestamp,
    deleted_at               timestamptz,
    finalizers               text        not null default '{}',

    reconcile_status         text        not null default 'Initializing',
    reconcile_status_details text        not null default '',

    name                     text        not null
);

create table workspace_access
(
    workspace_id text not null references workspace (id) on delete cascade,
    user_id      text not null references "user" (id) on delete restrict,

    primary key (workspace_id, user_id)
);

create table workspace_quotas
(
    workspace_id  text not null primary key references workspace (id) on delete cascade,

    max_log_bytes int  not null default 100
);
