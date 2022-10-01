create table if not exists users (
    id bigserial primary key,
    created_at timestamp(0) with time zone not null default NOW(),
    name text not null,
    email text unique not null,
    password_hash bytea not null,
    activated bool not null,
    version integer not null default 1
)