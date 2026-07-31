drop trigger if exists set_updated_at on users;
drop trigger if exists update_orders_updated_at on order_items;
drop function if exists update_order_updated_at();