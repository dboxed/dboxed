create table "user"
(
    id         text        not null primary key,
    created_at timestamptz not null default current_timestamp,

    username   text,

    email      text,
    full_name  text,
    avatar     text
);
