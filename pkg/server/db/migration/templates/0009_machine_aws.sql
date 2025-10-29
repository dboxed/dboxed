create table machine_aws
(
    id                       bigint not null primary key references machine (id) on delete cascade,

    reconcile_status         text   not null default 'Initializing',
    reconcile_status_details text   not null default '',

    instance_type            text   not null,
    subnet_id                text   not null,
    root_volume_size         bigint not null
);

create table machine_aws_status
(
    id          bigint not null primary key references machine (id) on delete cascade,

    instance_id text unique
);
