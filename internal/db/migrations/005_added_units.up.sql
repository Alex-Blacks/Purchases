create table units(
    id integer generated always as identity primary key,
    name varchar(50) unique not null,
    short_name varchar(50) unique not null    
);

alter table order_items add column unit_id integer;

alter table order_items
    add constraint fk_order_items_unit_id
    foreign key (unit_id) references units (id) on delete restrict;

create index idx_order_items_unit_id on order_items (unit_id);