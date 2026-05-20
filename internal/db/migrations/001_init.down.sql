drop table if exists purchases cascade;
drop table if exists purchase_items;
drop table if exists orders cascade;
drop table if exists order_items;
drop table if exists products cascade;
drop table if exists product_aliases;
drop table if exists categories;
drop table if exists stores;
drop table if exists users;

drop extension if exists citext;
drop type if exists user_status;
drop type if exists user_role;