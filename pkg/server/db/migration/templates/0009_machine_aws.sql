create table machine_aws
(
    id               bigint primary key references machine (id) on delete cascade,
    instance_type    text   not null,
    subnet_id        text   not null,
    root_volume_size bigint not null
);

create table machine_aws_status
(
    id          bigint primary key references machine (id) on delete cascade,

    instance_id text unique
);
