drop index if exists idx_history_created_at;
drop index if exists idx_history_action;
drop index if exists idx_history_entity_time;
drop index if exists idx_history_group;

alter table change_history drop constraint if exists chk_history_data;

drop table if exists change_history cascade;

drop type if exists history_action;
drop type if exists history_entity;
