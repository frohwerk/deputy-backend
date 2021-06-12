-- History for deployments table
CREATE OR REPLACE FUNCTION write_deployments_history() RETURNS trigger AS $$
  BEGIN
    CASE TG_OP
      WHEN 'INSERT' THEN
        NEW.updated := CURRENT_TIMESTAMP;
        RETURN NEW;
      WHEN 'UPDATE' THEN
        NEW.updated := CURRENT_TIMESTAMP;
        INSERT INTO deployments_history (component_id, platform_id, valid_from, valid_until, image_ref)
        VALUES(OLD.component_id, OLD.platform_id, OLD.updated, NEW.updated, OLD.image_ref);
        RETURN NEW;
      WHEN 'DELETE' THEN
        INSERT INTO deployments_history (component_id, platform_id, valid_from, valid_until, image_ref)
        VALUES(OLD.component_id, OLD.platform_id, OLD.updated, CURRENT_TIMESTAMP, OLD.image_ref);
        RETURN OLD;
    END CASE;
  END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER write_deployments_history
BEFORE INSERT OR UPDATE OR DELETE ON deployments
FOR EACH ROW EXECUTE FUNCTION write_deployments_history();

-- Change notifications for all tables?
CREATE OR REPLACE FUNCTION deployments_notify() RETURNS trigger AS $$
  BEGIN
    CASE TG_OP
      WHEN 'INSERT' THEN PERFORM pg_notify('demo_channel', FORMAT('%s;%s;%s;%s', TG_OP, NEW.component_id, NEW.platform_id, NEW.updated));
      WHEN 'UPDATE' THEN PERFORM pg_notify('demo_channel', FORMAT('%s;%s;%s;%s', TG_OP, NEW.component_id, NEW.platform_id, NEW.updated));
      WHEN 'DELETE' THEN PERFORM pg_notify('demo_channel', FORMAT('%s;%s;%s;%s', TG_OP, OLD.component_id, OLD.platform_id, OLD.updated));
      ELSE RAISE EXCEPTION 'Unsupported operation in trigger function deployments_notify: %', TG_OP;
    END CASE;
    RETURN NEW;
  END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER deployments_notify_trigger
AFTER INSERT OR UPDATE OR DELETE ON deployments
FOR EACH ROW
EXECUTE FUNCTION deployments_notify();
