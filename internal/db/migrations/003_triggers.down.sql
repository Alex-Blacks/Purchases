drop trigger if exists update_order_on_change on order_items;
drop trigger if exists expires_invite on invites;
drop trigger if exists update_users_updated_at on users;

drop function if exists update_order_updated_at();
drop function if exists set_invite_expires();
drop function if exists set_updated_at();
