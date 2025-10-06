create table workspace
(
    id                       TYPES_INT_PRIMARY_KEY,
    created_at               TYPES_DATETIME not null default current_timestamp,
    deleted_at               TYPES_DATETIME,
    finalizers               text           not null default '{}',

    reconcile_status         text           not null default 'Initializing',
    reconcile_status_details text           not null default '',

    name                     text           not null
);

create table workspace_access
(
    workspace_id bigint not null references workspace (id) on delete cascade,
    user_id      text   not null references "user" (id) on delete restrict,

    primary key (workspace_id, user_id)
);
