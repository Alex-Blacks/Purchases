create type history_entity as enum('store', 'unit', 'product', 'productAlias ', 'order', 'orderItem');
create type history_action as enum('create', 'update', 'delete');

create table change_history(
    id integer generated always as identity primary key,
    group_id integer not null references groups(id) on delete cascade,
    user_id integer not null references users(id) on delete cascade,
    entity_type history_entity not null,
    entity_id integer not null,
    action history_action not null,
    old_data jsonb,
    new_data jsonb,
    created_at timestamptz not null default now()
);

alter table change_history add CONSTRAINT chk_history_data CHECK (
    (action = 'create' and old_data is null and new_data is not null) or
    (action = 'delete' and old_data is not null and new_data is null) or
    (action = 'update' and old_data is not null and new_data is not null)
);

create index idx_history_group on change_history(group_id);
create index idx_history_entity_time on change_history(entity_type, entity_id, created_at desc);
create index idx_history_action on change_history(action);
create index idx_history_created_at on change_history(created_at);
create index idx_history_group_created on change_history(group_id, created_at desc);
create index idx_history_user on change_history(user_id);