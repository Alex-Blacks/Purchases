create type user_role as enum('user', 'admin');

create type user_status as enum('active', 'blocked');

create type product_unit as enum('кг', 'г', 'мг', 'л', 'мл', 'шт');

create table users(
    id integer generated always as identity primary key,
    name varchar(50) not null,
    password varchar(100) not null,
    email varchar(40) unique not null,
    role user_role default 'user',
    status user_status default 'active',
    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

create table groups(
    id integer generated always as identity primary key,
    name varchar(50) unique not null
);

create table products(
    id integer generated always as identity primary key,
    title varchar(50) not null,
    unit product_unit not null,
    group_id integer not null,
    foreign key(group_id) references groups(id)
);

create table stores(
    id integer generated always as identity primary key,
    name varchar(50) unique not null
);

create table purchases(
    id integer generated always as identity primary key,
    user_id integer not null,
    store_id integer not null,
    product_id integer not null,
    count integer check (count>0) default 1,
    added_at timestamptz default now(),
    foreign key(user_id) references users(id),
    foreign key(store_id) references stores(id),
    foreign key(product_id) references products(id)
);

create table history(
    id integer generated always as identity primary key,
    user_id integer not null,
    store_id integer not null,
    product_id integer not null,
    count integer check (count>0) default 1,
    price numeric(10,2) check (price>0) not null,
    date timestamptz not null,
    foreign key(user_id) references users(id),
    foreign key(store_id) references stores(id),
    foreign key(product_id) references products(id)
);