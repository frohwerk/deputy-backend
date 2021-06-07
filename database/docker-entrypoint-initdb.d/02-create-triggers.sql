-- History for deployments table
CREATE OR REPLACE FUNCTION deployment_history_update() RETURNS trigger AS $$
  BEGIN
    NEW.updated := CURRENT_TIMESTAMP;
    INSERT INTO deployments_history (component_id, platform_id, valid_from, valid_until, image_ref)
    VALUES(OLD.component_id, OLD.platform_id, OLD.updated, NEW.updated, OLD.image_ref);
    RETURN NEW;
  END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER on_deployments_update
BEFORE UPDATE ON deployments
FOR EACH ROW
EXECUTE PROCEDURE deployment_history_update();

-- SELECT CONCAT(NEW.component_id, ';', NEW.platform_id';', NEW.valid_from) INTO payload;
-- payload := CONCAT('Hallo', ' ', 'Welt', '!');

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

DROP TRIGGER deployments_notify_trigger ON deployments;
CREATE TRIGGER deployments_notify_trigger
--BEFORE INSERT OR UPDATE OR DELETE ON deployments
AFTER INSERT OR UPDATE OR DELETE ON deployments
FOR EACH ROW
EXECUTE FUNCTION deployments_notify();

SELECT * FROM deployments;
UPDATE deployments SET updated = current_timestamp WHERE image_ref = 'registry.redhat.io/rhel8/postgresql-12';
