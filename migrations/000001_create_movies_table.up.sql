create table if not exists movies (
    id bigserial PRIMARY KEY, 
    create_at timestamp(0) with time zone not null default now(),
    title text not null,
    year integer not null,
    runtime integer not null,
    genres text[] not null,
    version integer not null default 1
)