create table git_spec
(
    id                       text        not null primary key,
    workspace_id             text        not null references workspace (id) on delete cascade,
    created_at               timestamptz not null default current_timestamp,
    deleted_at               timestamptz,
    finalizers               text        not null default '{}',

    change_seq               bigint      not null,
    reconcile_status         text        not null default 'Initializing',
    reconcile_status_details text        not null default '',

    git_url                  text        not null,
    git_ref                  text,
    subdir                   text        not null,
    spec_file                text        not null
);
