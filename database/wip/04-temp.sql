SELECT * FROM draft.apps;
SELECT * FROM draft.apps_timeline;
SELECT components.*, valid_from, valid_until FROM draft.apps JOIN draft.apps_components_all r ON app_id = apps.id JOIN draft.components ON components.id = component_id ORDER BY component_id, valid_from;
SELECT * FROM draft.deployments_all ORDER BY component_id, valid_from;

DROP VIEW draft.vapps_components;
CREATE VIEW draft.vapps_components AS (
  SELECT apps.id as app_id, apps.name as app_name,
         components.id as component_id, components.name as component_name,
         valid_from, valid_until
    FROM draft.apps
    JOIN draft.apps_components_all r ON app_id = apps.id
    JOIN draft.components ON components.id = component_id
);

SELECT * FROM draft.vapps_components ORDER BY component_id, valid_from;

-- This is it! Heureka!
-- Look no further, if you are wondering how to construct a timeline from subsequently dependent tables
-- The apps_timeline table could probably be substituted by a CTE, but I am unsure about the performance...
SELECT * FROM draft.apps_timeline;
SELECT * FROM draft.apps_components_all;
SELECT * FROM draft.deployments_all;
SELECT * FROM draft.apps_history;

CREATE VIEW draft.apps_history AS
SELECT t.app_id, t.iteration, t.valid_from, t.valid_until, c.component_id, d.image_ref --, d.valid_from, d.valid_until
  FROM draft.apps_timeline t
 INNER JOIN draft.apps_components_all c
    ON c.app_id = t.app_id
   AND c.valid_from <= t.valid_from AND t.valid_from < COALESCE(c.valid_until, CURRENT_TIMESTAMP)
  LEFT JOIN draft.deployments_all d
    ON d.component_id = c.component_id
   AND d.valid_from <= t.valid_from AND t.valid_from < COALESCE(d.valid_until, CURRENT_TIMESTAMP)
 WHERE t.app_id = 'demo'
 ORDER BY t.valid_from, c.component_id, d.valid_from;

-- History only for membership...
SELECT t.*, c.component_id, c.valid_from, c.valid_until
  FROM draft.apps_timeline t
 INNER JOIN draft.apps_components_all c
    ON c.app_id = t.app_id
   AND c.valid_from <= t.valid_from AND t.valid_from < COALESCE(c.valid_until, CURRENT_TIMESTAMP)
 WHERE t.app_id = 'demo'
 ORDER BY t.valid_from, c.component_id
;

BETWEEN '2021-06-03 12:30:41.141' AND '2021-06-03 12:32:54.444'

WITH
apps_history AS (
  SELECT valid_from FROM draft.apps_components_all UNION SELECT valid_until as valid_from FROM draft.apps_components_all
)
SELECT * from apps_history ORDER BY valid_from;
SELECT * FROM draft.vapps_components WHERE app_id = 'demo' ORDER BY valid_from, component_id;

SELECT * FROM draft.apps;
SELECT components.*, valid_from, valid_until FROM draft.apps JOIN draft.apps_components_all r ON app_id = apps.id JOIN draft.components ON components.id = component_id ORDER BY component_id, valid_from;
SELECT * FROM draft.deployments_all ORDER BY component_id, valid_from;

WITH
apps_history AS (
  SELECT valid_from FROM draft.apps_components_all UNION SELECT valid_until as valid_from FROM draft.apps_components_all
  UNION
  SELECT valid_from FROM draft.deployments_all UNION SELECT valid_until as valid_from FROM draft.deployments_all
)
SELECT * from apps_history ORDER BY valid_from;

--apps AS (
--  SELECT id, name FROM draft.apps
--),
--components AS (
--  SELECT id, name, valid_from, valid_until FROM draft.apps_components_all WHERE app_id = apps.id
--),


-- IDEE:
SELECT max(apps_components_all.valid_from, deployments_all.valid_from), min(apps_components_all.valid_to, deployments_all.valid_to)
-- usw...

CREATE OR REPLACE FUNCTION draft.wip(_app_id VARCHAR(36)) RETURNS RECORD AS $$
  DECLARE
    result record;
    x record;
    y record;
  BEGIN
    RAISE NOTICE 'function wip called for app %', _app_id;
    FOR x IN SELECT envs.id, envs.name FROM draft.envs LOOP
      RAISE NOTICE 'env % %', x.id, x.name;
      FOR y IN SELECT platforms.id, platforms.name FROM draft.platforms LOOP
        RAISE NOTICE 'platform % %', y.id, y.name;
      END LOOP;
      FOR y IN
        WITH
          c AS (
            SELECT apps_components.app_id as app_id, json_build_object(components.name, json_build_object('image', deployments.image_ref, 'updated', deployments.updated)) as _components
              FROM draft.apps_components
             INNER JOIN draft.components ON components.id = apps_components.component_id
              LEFT JOIN draft.deployments ON deployments.component_id = components.id
             WHERE deployments.platform_id in (SELECT platforms.id FROM draft.envs JOIN draft.platforms ON platforms.env_id = envs.id)
          )
        --SELECT json_build_object('id', apps.id, 'name', apps.name, 'components', _components) AS result
        SELECT apps.id, apps.name, _components
          FROM draft.apps INNER JOIN c on c.app_id = apps.id
         WHERE apps.id = _app_id
      LOOP
        RAISE NOTICE 'result % % %', y.id, y.name, y._components;
      END LOOP;
    END LOOP;
    RETURN result;
  END;
$$ LANGUAGE plpgsql;
SELECT draft.wip('demo');
