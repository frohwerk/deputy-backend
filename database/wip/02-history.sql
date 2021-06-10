CREATE TABLE apps_history (
    app_id VARCHAR(36) NOT NULL,
    env_id VARCHAR(36) NOT NULL,
    valid_from TIMESTAMP NOT NULL,
    components JSON NOT NULL,
    PRIMARY KEY (app_id, env_id, valid_from),
    FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE,
    FOREIGN KEY (app_id) REFERENCES envs(env_id) ON DELETE CASCADE
);

SELECT DISTINCT apps.id, apps.name FROM deployments JOIN apps_components ON  =  WHERE apps_components.component_id = '';

SELECT *
  FROM apps
 CROSS JOIN envs
;

-- pseudocode
-- whenever deployments changes:
--    for each env:
--      write components snapshot of app/env
-- END pseudocode
-- Done so far: Create a snapshot for a specific app_id
-- TODO: How to handle a missing deployment?
CREATE OR REPLACE FUNCTION wip(_app_id VARCHAR(36)) RETURNS JSON AS $$
  DECLARE
    result JSON;
  BEGIN
    RAISE NOTICE 'function wip called for app %', _app_id;
    WITH
      c AS (
        SELECT apps_components.app_id as app_id, json_build_object(components.name, json_build_object('image', deployments.image_ref, 'updated', deployments.updated)) as _components
          FROM apps_components
         INNER JOIN components ON components.component_id = apps_components.component_id
          LEFT JOIN deployments ON deployments.component_id = components.component_id
      )
    SELECT json_build_object('id', apps.id, 'name', apps.name, 'components', _components)
      INTO result
      FROM apps INNER JOIN c on c.app_id = apps.id
     WHERE apps.id = _app_id;
    RETURN result;
  END;
$$ LANGUAGE plpgsql;
SELECT wip('383b3b6e-9094-4b52-b571-4e520391d125');

SELECT to_json(c.*) FROM components c WHERE component_id = '954b818b-7295-4b4d-8f4b-eb120119ac0a';

-- Update apps_history on deployments changes
CREATE OR REPLACE FUNCTION write_deployments_history() RETURNS trigger AS $$
  DECLARE
    component components%ROWTYPE;
    app_id VARCHAR(36);
    app_name VARCHAR(36);
    component_id VARCHAR(36);
    image_ref VARCHAR(150);
    mod_timestamp := CURRENT_TIMESTAMP;
  BEGIN
    CASE TG_OP
      WHEN 'INSERT' THEN
        component = NEW;
      WHEN 'UPDATE' THEN
        component = NEW;
      WHEN 'DELETE' THEN
        component = OLD;
    END CASE;
    SELECT id, name INTO app_id, app_name FROM apps_components JOIN apps ON apps.id = apps_components.app_id WHERE apps_components.component_id = component.component_id;
    FOR r IN SELECT  FROM 
    SELECT * FROM apps
      JOIN apps_components on apps_components.app_id = apps.id
      JOIN components on components.component_id = apps_components.component_id
     WHERE components.component_id = OLD.
  END;
$$ LANGUAGE plpgsql;
