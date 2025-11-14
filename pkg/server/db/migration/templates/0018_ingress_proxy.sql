create table ingress_proxy
(
    id                       TYPES_UUID_PRIMARY_KEY,
    workspace_id             text           not null references workspace (id) on delete cascade,
    created_at               TYPES_DATETIME not null default current_timestamp,
    deleted_at               TYPES_DATETIME,
    finalizers               text           not null default '{}',

    reconcile_status         text           not null default 'Initializing',
    reconcile_status_details text           not null default '',

    name                     text           not null,
    proxy_type               text           not null,
    network_id               text           not null references network (id) on delete restrict,
    replicas                 int            not null,

    http_port                int            not null,
    https_port               int            not null,

    unique (workspace_id, name)
);

create table ingress_proxy_box
(
    ingress_proxy_id text not null references ingress_proxy (id) on delete cascade,
    box_id           text not null references box (id) on delete cascade
);

create table box_ingress
(
    id          TYPES_UUID_PRIMARY_KEY,
    created_at  TYPES_DATETIME not null default current_timestamp,

    proxy_id    text           not null references ingress_proxy (id) on delete restrict,

    box_id      text           not null references box (id) on delete cascade,
    description text,

    hostname    text           not null,
    path_prefix text           not null,
    port        int            not null
);
