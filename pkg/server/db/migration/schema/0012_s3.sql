create table s3_bucket
(
    id                       text not null primary key,
    workspace_id             text           not null references workspace (id) on delete restrict,
    created_at               timestamptz not null default current_timestamp,
    deleted_at               timestamptz,
    finalizers               text           not null default '{}',

    reconcile_status         text           not null default 'Ok',
    reconcile_status_details text           not null default '',

    endpoint                 text           not null,
    bucket                   text           not null,
    access_key_id            text           not null,
    secret_access_key        text           not null
);

create index s3_bucket_workspace_bucket on s3_bucket (workspace_id, bucket);
