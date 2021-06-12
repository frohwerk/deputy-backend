-- History for deployments table
CREATE OR REPLACE FUNCTION write_deployments_history() RETURNS trigger AS $$
  DECLARE
    app apps.id%TYPE;
    env envs.id%TYPE;
    _timestamp  timestamp = CURRENT_TIMESTAMP;
  BEGIN
    CASE TG_OP

      WHEN 'INSERT' THEN
        NEW.updated := COALESCE(NEW.updated, _timestamp);
        FOR app IN
          SELECT DISTINCT app_id FROM apps_components WHERE component_id = NEW.component_id
        LOOP
          INSERT INTO apps_timeline (app_id, valid_from) VALUES (app, NEW.updated) ON CONFLICT DO NOTHING;
        END LOOP;
        RETURN NEW;

      WHEN 'UPDATE' THEN
        RAISE NOTICE 'update deployments: %, %, %, %', OLD.component_id, OLD.platform_id, NEW.updated, NEW.image_ref;
        CASE
          WHEN NEW.image_ref = OLD.image_ref THEN RETURN NEW;
          WHEN NEW.updated = OLD.updated OR NEW.updated IS NULL THEN NEW.updated = _timestamp;
        END CASE;
        INSERT INTO deployments_history (component_id, platform_id, valid_from, valid_until, image_ref)
        VALUES(OLD.component_id, OLD.platform_id, OLD.updated, NEW.updated, OLD.image_ref);
        FOR app IN
          SELECT DISTINCT app_id FROM apps_components WHERE component_id = OLD.component_id
        LOOP
          RAISE NOTICE 'insert into apps_timeline: %, %', app, NEW.updated;
          INSERT INTO apps_timeline (app_id, valid_from) VALUES (app, NEW.updated) ON CONFLICT DO NOTHING;
        END LOOP;
        RETURN NEW;

      WHEN 'DELETE' THEN
        INSERT INTO deployments_history (component_id, platform_id, valid_from, valid_until, image_ref)
        VALUES(OLD.component_id, OLD.platform_id, OLD.updated, CURRENT_TIMESTAMP, OLD.image_ref);
        FOR app IN
          SELECT DISTINCT app_id FROM apps_components WHERE component_id = OLD.component_id
        LOOP
          INSERT INTO apps_timeline (app_id, valid_from) VALUES (app, OLD.updated) ON CONFLICT DO NOTHING;
        END LOOP;
        RETURN OLD;

    END CASE;
  END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER write_deployments_history
BEFORE INSERT OR UPDATE OR DELETE ON deployments
FOR EACH ROW EXECUTE FUNCTION write_deployments_history();

-- History for apps_components table
CREATE OR REPLACE FUNCTION write_apps_components_history() RETURNS trigger AS $$
  DECLARE
    _env         envs.id%TYPE;
    _timestamp  timestamp = CURRENT_TIMESTAMP;
  BEGIN
    RAISE NOTICE 'trigger write_apps_components_history: %', TG_OP;
    CASE TG_OP
      WHEN 'INSERT' THEN
        NEW.updated := COALESCE(NEW.updated, _timestamp);
        RAISE NOTICE 'apps_components: %, %, %', NEW.app_id, NEW.component_id, NEW.updated;
        INSERT INTO apps_timeline (app_id, valid_from) VALUES (NEW.app_id, _timestamp) ON CONFLICT DO NOTHING;
        RETURN NEW;
      WHEN 'UPDATE' THEN
        -- TODO: Do not allow updates, the relationship is stateless (except for the modification timestamp)
        RAISE NOTICE 'invalid modification: update on apps_components: %, %', OLD.app_id, OLD.component_id;
        NEW.updated := COALESCE(NEW.updated, _timestamp);
        RETURN NEW;
      WHEN 'DELETE' THEN
        INSERT INTO apps_components_history (app_id, component_id, valid_from, valid_until)
        VALUES(OLD.app_id, OLD.component_id, OLD.updated, _timestamp);
        INSERT INTO apps_timeline (app_id, valid_from) VALUES (OLD.app_id, _timestamp) ON CONFLICT DO NOTHING;
        RETURN OLD;
    END CASE;
  END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER write_apps_components_history
BEFORE INSERT OR UPDATE OR DELETE ON apps_components
FOR EACH ROW EXECUTE FUNCTION write_apps_components_history();

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
