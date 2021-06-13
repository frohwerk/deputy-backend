-- Debug triggers for writing to apps_timeline
-- CREATE OR REPLACE FUNCTION deployments_print_changes() RETURNS void AS $$

-- Helper function for writing to apps_timeline
CREATE OR REPLACE FUNCTION write_apps_timeline(_platform_id CHARACTER VARYING, _component_id CHARACTER VARYING, _valid_from TIMESTAMP) RETURNS void AS $$
  BEGIN
    RAISE NOTICE 'write_apps_timeline(%, %, %)', _platform_id, _component_id, _valid_from;
    INSERT INTO apps_timeline (app_id, env_id, valid_from)
    SELECT app_id, env_id, _valid_from FROM apps_components CROSS JOIN platforms
     WHERE platforms.id = _platform_id AND apps_components.component_id = _component_id
        ON CONFLICT DO NOTHING;
  END;
$$ LANGUAGE plpgsql;
-- History for deployments table
CREATE OR REPLACE FUNCTION write_deployments_history() RETURNS trigger AS $$
  DECLARE
    _timestamp  timestamp = CURRENT_TIMESTAMP;
  BEGIN
    CASE TG_OP

      WHEN 'INSERT' THEN
        NEW.updated := COALESCE(NEW.updated, _timestamp);
        PERFORM write_apps_timeline(NEW.platform_id, NEW.component_id, NEW.updated);
        RETURN NEW;

      WHEN 'UPDATE' THEN
        RAISE NOTICE 'update deployments: %, %, %, %', OLD.component_id, OLD.platform_id, NEW.updated, NEW.image_ref;
        CASE
          WHEN NEW.image_ref = OLD.image_ref THEN RETURN NEW;
          WHEN NEW.updated = OLD.updated OR NEW.updated IS NULL THEN NEW.updated = _timestamp;
        END CASE;
        INSERT INTO deployments_history (component_id, platform_id, valid_from, valid_until, image_ref)
               VALUES(OLD.component_id, OLD.platform_id, OLD.updated, NEW.updated, OLD.image_ref)
               ON CONFLICT DO NOTHING;
        PERFORM write_apps_timeline(OLD.platform_id, OLD.component_id, NEW.updated);
        RETURN NEW;

      WHEN 'DELETE' THEN
        RAISE NOTICE 'delete deployments: %, %, %, %', OLD.component_id, OLD.platform_id, OLD.updated, OLD.image_ref;
        RAISE NOTICE 'TG_TABLE_NAME: %', TG_TABLE_NAME;
        INSERT INTO deployments_history (component_id, platform_id, valid_from, valid_until, image_ref)
               SELECT components.id, platforms.id, OLD.updated, _timestamp, OLD.image_ref
               FROM components CROSS JOIN platforms WHERE components.id = OLD.component_id AND platforms.id = OLD.platform_id
               ON CONFLICT DO NOTHING;
        PERFORM write_apps_timeline(OLD.platform_id, OLD.component_id, _timestamp);
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
        INSERT INTO apps_timeline (app_id, env_id, valid_from) SELECT NEW.app_id, envs.id, _timestamp FROM envs ON CONFLICT DO NOTHING;
        RETURN NEW;
      WHEN 'UPDATE' THEN
        -- TODO: Do not allow updates, the relationship is stateless (except for the modification timestamp)
        RAISE NOTICE 'invalid modification: update on apps_components: %, %', OLD.app_id, OLD.component_id;
        NEW.updated := COALESCE(NEW.updated, _timestamp);
        RETURN NEW;
      WHEN 'DELETE' THEN
        INSERT INTO apps_components_history (app_id, component_id, valid_from, valid_until)
               SELECT apps.id, components.id, OLD.updated, _timestamp
               FROM apps CROSS JOIN components WHERE apps.id = OLD.app_id AND components.id = OLD.component_id
               ON CONFLICT DO NOTHING;
        INSERT INTO apps_timeline (app_id, env_id, valid_from)
               SELECT apps.id, envs.id, _timestamp
               FROM apps CROSS JOIN envs WHERE apps.id = OLD.app_id
               ON CONFLICT DO NOTHING;
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
