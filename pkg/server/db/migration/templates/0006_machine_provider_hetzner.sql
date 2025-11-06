create table machine_provider_hetzner
(
    id                   text not null primary key references machine_provider (id) on delete cascade,
    hcloud_token         text not null,

    hetzner_network_name text not null,

    robot_user           text,
    robot_password       text
);

create table machine_provider_hetzner_status
(
    id                   text not null primary key references machine_provider (id) on delete cascade,
    hetzner_network_id   bigint,
    hetzner_network_zone text,
    hetzner_network_cidr text,
    cloud_subnet_cidr    text,
    robot_subnet_cidr    text,
    robot_vswitch_id     bigint
);
