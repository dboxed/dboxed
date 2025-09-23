create table volume_provider_dboxed
(
    id            bigint primary key references volume_provider (id) on delete cascade,

    api_url       text   not null,
    token         text   not null,
    repository_id bigint not null
);

create table volume_provider_dboxed_status
(
    id        bigint primary key references volume_provider (id) on delete cascade,

    volume_id bigint,
    fs_size   text,
    fs_type   text
);
