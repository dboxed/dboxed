create table "user"
(
    id         text not null primary key,
    created_at TYPES_DATETIME not null default current_timestamp,

    name       text not null,
    email      text,
    avatar     text
);

