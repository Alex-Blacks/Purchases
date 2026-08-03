CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN
    new.updated_at = NOW();
    RETURN new;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_order_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN

    IF EXISTS (SELECT 1 FROM orders WHERE id = OLD.order_id) THEN
        UPDATE orders SET updated_at = NOW() WHERE id = OLD.order_id;
    END IF;
    RETURN OLD;
    ELSE
        UPDATE orders SET updated_at = NOW() WHERE id = NEW.order_id;
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_order_on_change
    AFTER INSERT OR UPDATE OR DELETE ON order_items
    FOR EACH ROW
    EXECUTE FUNCTION update_order_updated_at();


CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();