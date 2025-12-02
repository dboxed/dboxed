create table machine_hetzner
(
    id                       text not null primary key references machine (id) on delete cascade,

    reconcile_status         text not null default 'Initializing',
    reconcile_status_details text not null default '',

    server_type              text not null,
    server_location          text not null
);

create table machine_hetzner_status
(
    id        text not null primary key references machine (id) on delete cascade,

    server_id bigint
);

