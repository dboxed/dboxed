create table git_credentials
(
    id               text        not null primary key,
    workspace_id     text        not null references workspace (id) on delete cascade,
    created_at       timestamptz not null default current_timestamp,

    host             text        not null,
    path_glob        text        not null,

    credentials_type text        not null,
    username         text,
    password         text,
    ssh_key          text,

    unique (workspace_id, host, path_glob)
);
