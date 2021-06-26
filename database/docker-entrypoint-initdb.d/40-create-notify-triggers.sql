CREATE OR REPLACE FUNCTION change_notify() RETURNS trigger AS $$
  BEGIN
    CASE TG_OP
      WHEN 'INSERT' THEN
        PERFORM pg_notify(TG_TABLE_NAME, FORMAT('%s:%s', SUBSTR(TG_OP, 1, 1), NEW.ID));
        RETURN NEW;
      WHEN 'UPDATE' THEN
        PERFORM pg_notify(TG_TABLE_NAME, FORMAT('%s:%s', SUBSTR(TG_OP, 1, 1), NEW.ID));
        RETURN NEW;
      WHEN 'DELETE' THEN
        PERFORM pg_notify(TG_TABLE_NAME, FORMAT('%s:%s', SUBSTR(TG_OP, 1, 1), OLD.ID));
        RETURN OLD;
      ELSE
        RAISE NOTICE 'UNEXPECTED TG_OP: %', TG_OP;
    END CASE;
    RETURN NEW;
  END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER change_notify AFTER INSERT OR UPDATE OR DELETE ON platforms FOR EACH ROW EXECUTE FUNCTION change_notify();
-----------------------------------------------------------------------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION deploy_notify() RETURNS trigger AS $$
  BEGIN
    CASE TG_OP
      WHEN 'INSERT' THEN
        PERFORM pg_notify('images', NEW.IMAGE_REF);
        RETURN NEW;
      WHEN 'UPDATE' THEN
        PERFORM pg_notify('images', NEW.IMAGE_REF);
        RETURN NEW;
      WHEN 'DELETE' THEN
        RETURN OLD;
      ELSE
        RAISE NOTICE 'UNEXPECTED TG_OP: %', TG_OP;
    END CASE;
    RETURN NEW;
  END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER deploy_notify AFTER INSERT OR UPDATE ON deployments FOR EACH ROW EXECUTE FUNCTION deploy_notify();
-----------------------------------------------------------------------------------------------------------------------------------------------------------------
