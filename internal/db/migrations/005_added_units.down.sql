drop index if exists idx_order_items_unit_id;

alter table order_items drop constraint if exists fk_order_items_unit_id;
alter table order_items drop COLUMN if exists unit_id;

drop table if exists units;
