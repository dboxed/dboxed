create table volume_snapshot
(
    id                 TYPES_UUID_PRIMARY_KEY,
    workspace_id       text           not null references workspace (id) on delete restrict,
    created_at         TYPES_DATETIME not null default current_timestamp,
    deleted_at         TYPES_DATETIME,
    finalizers         text           not null default '{}',

    volume_provider_id text           not null references volume_provider (id) on delete restrict,
    volume_id          text references volume (id) on delete restrict,

    mount_id           text           not null
);

--{{ if eq .DbType "postgres" }}
alter table volume
    add foreign key (latest_snapshot_id) references volume_snapshot (id) on delete restrict;
--{{ end }}

create table volume_snapshot_rustic
(
    id                      text           not null primary key references volume_snapshot (id) on delete cascade,

    snapshot_id             text           not null unique,
    snapshot_time           TYPES_DATETIME not null,
    parent_snapshot_id      text,

    hostname                text           not null,

    files_new               int            not null,
    files_changed           int            not null,
    files_unmodified        int            not null,
    total_files_processed   int            not null,
    total_bytes_processed   int            not null,
    dirs_new                int            not null,
    dirs_changed            int            not null,
    dirs_unmodified         int            not null,
    total_dirs_processed    int            not null,
    total_dirsize_processed int            not null,
    data_blobs              int            not null,
    tree_blobs              int            not null,
    data_added              int            not null,
    data_added_packed       int            not null,
    data_added_files        int            not null,
    data_added_files_packed int            not null,
    data_added_trees        int            not null,
    data_added_trees_packed int            not null,

    backup_start            TYPES_DATETIME not null,
    backup_end              TYPES_DATETIME not null,
    backup_duration         float4         not null,
    total_duration          float4         not null
);
