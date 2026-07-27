drop index if exists idx_order_items_order_product;

create unique index idx_unique_order_product ON order_items (order_id, product_id);