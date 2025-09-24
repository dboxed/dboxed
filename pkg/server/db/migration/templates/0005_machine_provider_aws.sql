create table machine_provider_aws
(
    id                    bigint not null primary key references machine_provider (id) on delete cascade,
    region                text   not null,
    aws_access_key_id     text,
    aws_secret_access_key text,
    vpc_id                text   not null
);

create table machine_provider_aws_status
(
    id                bigint not null primary key references machine_provider (id) on delete cascade,
    vpc_name          text,
    vpc_cidr          text,
    security_group_id text
);

create table machine_provider_aws_subnet
(
    machine_provider_id bigint not null references machine_provider (id) on delete cascade,
    subnet_id           text   not null,
    subnet_name         text,
    availability_zone   text   not null,
    cidr                text   not null,

    primary key (machine_provider_id, subnet_id)
);
