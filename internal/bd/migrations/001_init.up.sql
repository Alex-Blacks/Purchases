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

create table categories(
    id integer generated always as identity primary key,
    name varchar(50) unique not null
);

create table products(
    id integer generated always as identity primary key,
    title varchar(50) not null,
    unit product_unit not null,
    category_id integer not null,
    foreign key(category_id) references categories(id)
);

create table product_aliases(
    id integer generated always as identity primary key,
    product_id integer not null,
    alias varchar(50) not null,
    foreign key(product_id) references products(id)
);

create table stores(
    id integer generated always as identity primary key,
    name varchar(50) unique not null
);

create table orders(
    id integer generated always as identity primary key,
    user_id integer not null,
    store_id integer unique not null,
    created_at timestamptz default now(),
    updated_at timestamptz default now(),
    foreign key(user_id) references users(id),
    foreign key(store_id) references stores(id)
);

create table order_items(
    id integer generated always as identity primary key,
    order_id integer not null,
    product_id integer not null,
    quantity integer check (quantity > 0) default 1,
    foreign key(order_id) references orders(id),
    foreign key(product_id) references products(id)
);


create table purchases(
    id integer generated always as identity primary key,
    user_id integer not null,
    external_id integer,
    store_id integer not null,
    total_sum numeric(10,2) check (total_sum > 0) not null,
    purchased_at timestamptz default now(),
    raw_qr varchar(100) not null,
    foreign key(user_id) references users(id),
    foreign key(store_id) references stores(id)
);



create table purchase_items(
    id integer generated always as identity primary key,
    purchase_id integer not null,
    row_name varchar(100) not null,
    product_id integer null,
    quantity integer check (quantity > 0) default 1,
    price numeric(10,2) check (price > 0) not null,
    foreign key(purchase_id) references purchases(id),
    foreign key(product_id) references products(id) on delete set null
);