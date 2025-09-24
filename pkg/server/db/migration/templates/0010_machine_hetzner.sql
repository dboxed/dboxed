create table machine_hetzner
(
    id              bigint not null primary key references machine (id) on delete cascade,
    server_type     text   not null,
    server_location text   not null
);

create table machine_hetzner_status
(
    id        bigint not null primary key references machine (id) on delete cascade,

    server_id bigint
);

