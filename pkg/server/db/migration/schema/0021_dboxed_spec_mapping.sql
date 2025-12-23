create table dboxed_spec_mapping
(
    id            text        not null primary key,
    workspace_id  text        not null references workspace (id) on delete cascade,
    created_at    timestamptz not null default current_timestamp,

    spec_id       text        not null references dboxed_spec (id) on delete cascade,
    object_type   text        not null,
    object_id     text        not null,
    object_name   text        not null,

    recreate_key  text        not null,
    spec_fragment text        not null,

    unique (workspace_id, spec_id, object_type, object_name)
);
