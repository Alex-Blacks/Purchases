create type user_role as enum('user', 'admin');
create type user_status as enum('active', 'blocked');
create type invite_status as enum('pending','accepted','rejected');

create extension if not exists citext;

create table groups(
    id integer generated always as identity primary key,
    name varchar(50) not null,
    admin_user_id integer not null references users(id) on delete restrict
)

create table users(
    id integer generated always as identity primary key,
    name varchar(50) not null,
    password_hash text not null,
    email citext unique not null,
    group_id integer not null references groups(id) on delete cascade,
    role user_role not null default 'user',
    status user_status not null default 'active',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table invites(
    id integer generated always as identity primary key,
    group_id integer not null references groups(id) on delete cascade, 
    inviter_user_id integer not null references users(id) on delete cascade, 
    invitee_email citext not null, 
    status invate_status not null default 'pending', 
    token text not null unique, 
    created_at timestamptz not null default now(), 
    expires_at timestamptz not null
)

create table units(
    id integer generated always as identity primary key,
    name varchar(50) not null,
    short_name varchar(50) not null,
    group_id integer not null references groups(id) on delete cascade,
    constraint unique_unit_name_group unique (group_id, name),
    constraint unique_unit_short_name_group unique (group_id, short_name)
);


create table products(
    id integer generated always as identity primary key,
    title varchar(50) not null,
    group_id integer not null references groups(id) on delete cascade,
    constraint unique_product_title_group unique (group_id, title)
);

create table product_aliases(
    id integer generated always as identity primary key,
    product_id integer not null references products(id) on delete cascade,
    alias varchar(50) not null,
    group_id integer not null references groups(id) on delete cascade,
    constraint unique_product_alias unique (product_id, alias),
    constraint unique_product_alias_alias_group unique (group_id, alias)
);

create table stores(
    id integer generated always as identity primary key,
    name varchar(50) not null,
    group_id integer not null references groups(id) on delete cascade,
    constraint unique_store_name_group unique (group_id, name)
);

create table orders(
    id integer generated always as identity primary key,
    user_id integer not null references users(id) on delete cascade,
    store_id integer not null references stores(id) on delete restrict,
    group_id integer not null references groups(id) on delete cascade,
    created_at timestamptz default now(),
    updated_at timestamptz default now(),
    constraint uniq_group_store_order unique(group_id, store_id)
);

create table order_items(
    id integer generated always as identity primary key,
    order_id integer not null references orders(id) on delete cascade,
    product_id integer not null references products(id) on delete restrict,
    unit_id integer not null references units(id) on delete restrict,
    quantity numeric(10,3) check (quantity > 0) default 1,
    group_id integer not null references groups(id) on delete cascade
);


create table purchases(
    id integer generated always as identity primary key,
    user_id integer not null references users(id) on delete cascade,
    group_id integer not null references groups(id) on delete cascade,
    external_id integer,
    store_id integer null references stores(id) on delete set null,
    total_sum numeric(10,2) check (total_sum > 0) not null,
    purchased_at timestamptz default now(),
    raw_qr varchar(100) not null
);



create table purchase_items(
    id integer generated always as identity primary key,
    purchase_id integer not null references purchases(id) on delete cascade,
    row_name varchar(100) not null,
    product_id integer null references products(id) on delete set null,
    quantity numeric(10,3) check (quantity > 0) default 1,
    price numeric(10,2) check (price > 0) not null,
    group_id integer not null references groups(id) on delete cascade
);