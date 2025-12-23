create table machine_hetzner
(
    id                       text   not null primary key references machine (id) on delete cascade,

    change_seq               bigint not null,
    reconcile_status         text   not null default 'Initializing',
    reconcile_status_details text   not null default '',

    server_type              text   not null,
    server_location          text   not null
);
create index machine_hetzner_change_seq on machine_hetzner (change_seq);

create table machine_hetzner_status
(
    id         text not null primary key references machine (id) on delete cascade,

    server_id  bigint,
    public_ip4 text
);

