create table git_spec_mapping
(
    id           text        not null primary key,
    workspace_id text        not null references workspace (id) on delete cascade,
    created_at   timestamptz not null default current_timestamp,

    repo_key     text        not null,
    recreate_key text        not null,
    object_type  text        not null,
    object_id    text        not null,
    object_name  text        not null,

    spec         text        not null,

    --unique (workspace_id, repo_key, object_type, object_id),
    unique (workspace_id, repo_key, object_type, object_name)
);
