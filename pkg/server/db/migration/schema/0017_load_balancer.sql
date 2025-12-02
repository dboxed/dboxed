create table load_balancer
(
    id                       text        not null primary key,
    workspace_id             text        not null references workspace (id) on delete cascade,
    created_at               timestamptz not null default current_timestamp,
    deleted_at               timestamptz,
    finalizers               text        not null default '{}',

    reconcile_status         text        not null default 'Initializing',
    reconcile_status_details text        not null default '',

    name                     text        not null,
    load_balancer_type       text        not null,
    network_id               text        not null references network (id) on delete restrict,
    replicas                 int         not null,

    http_port                int         not null,
    https_port               int         not null,

    unique (workspace_id, name)
);

create table load_balancer_box
(
    load_balancer_id text not null references load_balancer (id) on delete cascade,
    box_id           text not null references box (id) on delete cascade
);

create table load_balancer_service
(
    id               text        not null primary key,
    created_at       timestamptz not null default current_timestamp,

    load_balancer_id text        not null references load_balancer (id) on delete restrict,

    box_id           text        not null references box (id) on delete cascade,
    description      text,

    hostname         text        not null,
    path_prefix      text        not null,
    port             int         not null
);

create table load_balancer_certmagic
(
    load_balancer_id text        not null references load_balancer (id) on delete cascade,
    key              text        not null,
    value            bytea       not null,
    last_modified    timestamptz not null,
    unique (load_balancer_id, key)
);
