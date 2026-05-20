create index idx_users_email on users(email);

create index idx_product_categories on products(category_id);

create index idx_product_aliases_products on product_aliases(product_id);
create unique index idx_product_aliases_unique_alias on product_aliases(product_id, lower(trim(alias)));

create index idx_orders_users on orders(user_id);
create index idx_orders_stores on orders(store_id);
create index idx_orders_user_id_id on orders(user_id, id);

create index idx_order_items_orders on order_items(order_id);
create index idx_order_items_products on order_items(product_id);
create index idx_order_items_order_product on order_items(order_id, product_id);

create index idx_purchases_users on purchases(user_id);
create index idx_purchases_stores on purchases(store_id);

create index idx_purchase_items_purchases on purchase_items(purchase_id);
create index idx_purchase_items_products on purchase_items(product_id);


