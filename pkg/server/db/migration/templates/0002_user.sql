create table "user"
(
    id         text           not null primary key,
    created_at TYPES_DATETIME not null default current_timestamp,

    name       text           not null,
    email      text,
    avatar     text
);

create table token
(
    id         TYPES_INT_PRIMARY_KEY,
    created_at TYPES_DATETIME not null default current_timestamp,

    token      text           not null unique,

    name       text           not null,
    user_id    text           not null references "user" (id) on delete cascade
);
