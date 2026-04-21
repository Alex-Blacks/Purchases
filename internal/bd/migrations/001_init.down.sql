drop table IF EXISTS purchases CASCADE;
drop table IF EXISTS purchase_items CASCADE;
drop table IF EXISTS orders CASCADE;
drop table IF EXISTS order_items CASCADE;
drop table IF EXISTS products CASCADE;
drop table IF EXISTS product_aliases CASCADE;
drop table IF EXISTS categories CASCADE;
drop table IF EXISTS stores CASCADE;
drop table IF EXISTS users CASCADE;

DROP TYPE IF EXISTS product_unit;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_role;