SELECT routine_catalog, routine_schema, routine_name, routine_type, external_language, routine_definition FROM information_schema.routines WHERE routine_schema = 'public';
SELECT trigger_name, event_manipulation, event_object_catalog, event_object_schema, event_object_table, action_timing, action_statement FROM information_schema.triggers WHERE trigger_schema = 'public';

DROP TABLE IF EXISTS draft.apps;
DROP TABLE IF EXISTS draft.apps_timeline;
DROP TABLE IF EXISTS draft.components;
DROP TABLE IF EXISTS draft.envs;
DROP TABLE IF EXISTS draft.platforms;
DROP TABLE IF EXISTS draft.apps_components_history;
DROP TABLE IF EXISTS draft.apps_components;
DROP TABLE IF EXISTS draft.deployments;
DROP TABLE IF EXISTS draft.deployments_history;

CREATE SCHEMA draft;

CREATE TABLE draft.apps (
    id   VARCHAR(36) NOT NULL DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE,
    PRIMARY KEY (id)
);

CREATE TABLE draft.apps_timeline (
  app_id       VARCHAR(36) NOT NULL,
  iteration    INTEGER NOT NULL,
  valid_from   TIMESTAMP NOT NULL,
  valid_until  TIMESTAMP,
  PRIMARY KEY (app_id, valid_from),
  FOREIGN KEY (app_id) REFERENCES draft.apps (id) ON DELETE CASCADE
);

CREATE TABLE draft.components (
    id   VARCHAR(36) NOT NULL DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    PRIMARY KEY (id)
);

CREATE TABLE draft.envs (
    id   VARCHAR(36) NOT NULL DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    PRIMARY KEY (id)
);

CREATE TABLE draft.platforms (
    id         VARCHAR(36) NOT NULL DEFAULT gen_random_uuid(),
    env_id     VARCHAR(36) NOT NULL,
    name       VARCHAR(50),
    api_server VARCHAR(256),
    namespace  VARCHAR(50),
    secret     VARCHAR(2048),
    PRIMARY KEY (id),
    UNIQUE (env_id, name),
    FOREIGN KEY (env_id) REFERENCES draft.envs(id) ON DELETE CASCADE
);

CREATE TABLE draft.apps_components (
    app_id VARCHAR(36) NOT NULL,
    component_id VARCHAR(36) NOT NULL,
    updated  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (app_id, component_id),
    FOREIGN KEY (app_id) REFERENCES draft.apps (id) ON DELETE CASCADE,
    FOREIGN KEY (component_id) REFERENCES draft.components (id) ON DELETE CASCADE
);

CREATE TABLE draft.apps_components_history (
    app_id VARCHAR(36) NOT NULL,
    component_id VARCHAR(36) NOT NULL,
    valid_from   TIMESTAMP NOT NULL,
    valid_until  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (app_id, component_id, valid_from),
    FOREIGN KEY (app_id) REFERENCES draft.apps (id) ON DELETE CASCADE,
    FOREIGN KEY (component_id) REFERENCES draft.components (id) ON DELETE CASCADE
);

CREATE TABLE draft.deployments (
    component_id VARCHAR(36) NOT NULL,
    platform_id  VARCHAR(36) NOT NULL,
    updated      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    image_ref    VARCHAR(150) NOT NULL,
    PRIMARY KEY (component_id, platform_id),
    FOREIGN KEY (platform_id) REFERENCES draft.platforms (id) ON DELETE CASCADE,
    FOREIGN KEY (component_id) REFERENCES draft.components (id) ON DELETE CASCADE
);

CREATE TABLE draft.deployments_history (
    component_id VARCHAR(36) NOT NULL,
    platform_id  VARCHAR(36) NOT NULL,
    valid_from   TIMESTAMP NOT NULL,
    valid_until  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    image_ref    VARCHAR(150) NOT NULL,
    PRIMARY KEY (component_id, platform_id, valid_from),
    FOREIGN KEY (platform_id) REFERENCES draft.platforms (id) ON DELETE CASCADE,
    FOREIGN KEY (component_id) REFERENCES draft.components (id) ON DELETE CASCADE
);

SELECT '2021-04-17 13:29:28.771'::timestamp;

DELETE FROM draft.apps;
INSERT INTO draft.apps (id, name) VALUES
  ('demo', 'Demo');

DELETE FROM draft.apps_timeline;
INSERT INTO draft.apps_timeline (app_id, iteration, valid_from, valid_until) VALUES
  ('demo', 0, '2021-06-03 11:30:24.114', '2021-06-03 12:30:41.141'),
  ('demo', 1, '2021-06-03 12:30:41.141', '2021-06-03 12:32:54.444'),
  ('demo', 2, '2021-06-03 12:32:54.444', '2021-06-05 09:11:28.871'),
  ('demo', 3, '2021-06-05 09:11:28.871', '2021-06-07 18:25:00.099'),
  ('demo', 4, '2021-06-07 18:25:00.099', '2021-06-08 13:29:28.771'),
  ('demo', 5, '2021-06-08 13:29:28.771', NULL);

DELETE FROM draft.components;
INSERT INTO draft.components (id, name) VALUES
  ('frontend', 'frontend'),
  ('demo-service', 'demo-service'),
  ('example-service', 'example-service'),
  ('auth-proxy', 'auth-proxy');

DELETE FROM draft.apps_components;
INSERT INTO draft.apps_components (app_id, component_id, updated) VALUES
  ('demo', 'frontend', '2021-06-03 12:32:54.444'),
  ('demo', 'demo-service', '2021-06-03 11:30:24.114'),
  ('demo', 'auth-proxy', '2021-06-08 13:29:28.771');

DELETE FROM draft.apps_components_history;
INSERT INTO draft.apps_components_history (app_id, component_id, valid_from, valid_until) VALUES
  ('demo', 'frontend', '2021-06-03 11:30:24.114', '2021-06-03 12:30:41.141'),
  ('demo', 'example-service', '2021-06-03 11:30:24.114', '2021-06-05 09:11:28.871');

DELETE FROM draft.envs;
INSERT INTO draft.envs (id, name) VALUES
  ('test', 'Test'),
  ('production', 'Produktion');

DELETE FROM draft.platforms;
INSERT INTO draft.platforms (id, env_id, name) VALUES
   ('test-inner', 'test', 'inner'),
   ('test-outer', 'test', 'outer'),
   ('prod-inner', 'production', 'inner'),
   ('prod-outer', 'production', 'outer');

DELETE FROM draft.deployments;
INSERT INTO draft.deployments (component_id, platform_id, updated, image_ref) VALUES
  ('frontend', 'test-inner', '2021-06-07 18:25:00.099', 'sample/frontend:1.0.4'),
  ('demo-service', 'test-outer', '2021-06-03 11:30:24.114', 'sample/demo-server:1.1.2'),
  ('example-service', 'test-outer', '2021-06-03 11:30:24.114', 'sample/example-server:1.0.1'),
  ('auth-proxy', 'test-inner', '2021-06-08 13:29:28.771', 'sample/oidc-proxy:1.4');

DELETE FROM draft.deployments_history;
INSERT INTO draft.deployments_history (component_id, platform_id, valid_from, valid_until, image_ref) VALUES
  ('frontend', 'test-inner', '2021-06-03 11:30:24.114', '2021-06-07 18:25:00.099', 'sample/frontend:1.0.3');

CREATE VIEW draft.apps_components_all AS
  SELECT app_id, component_id, updated as valid_from, null as valid_until FROM draft.apps_components
   UNION
  SELECT app_id, component_id, valid_from, valid_until FROM draft.apps_components_history;

CREATE VIEW draft.deployments_all AS
  SELECT component_id, platform_id, updated as valid_from, null as valid_until, image_ref FROM draft.deployments
   UNION
  SELECT component_id, platform_id, valid_from, valid_until, image_ref FROM draft.deployments_history;

CREATE OR REPLACE VIEW draft.result AS
SELECT apps.name as app_name,
       components.name as component_name, apps_components.valid_from as component_valid_from, apps_components.valid_until as component_valid_until,
       image_ref as image, deployments.valid_from as deployment_valid_from, deployments.valid_until as deployment_valid_until
  FROM draft.apps
 INNER JOIN draft.apps_components_all apps_components
    ON apps_components.app_id = apps.id
 INNER JOIN draft.components
    ON components.id = apps_components.component_id
  LEFT JOIN draft.deployments_all deployments
    ON deployments.component_id = components.id
 WHERE apps.id = 'demo'
 ORDER BY components.id, component_valid_from, deployment_valid_from
;
SELECT * FROM draft.result;

SELECT * FROM draft.apps_components;
SELECT * FROM draft.apps_components_history;
SELECT * FROM draft.deployments;
SELECT * FROM draft.deployments_history;

WITH apps_history AS (
  SELECT app_id, valid_from as updated FROM draft.apps_components_all
   UNION
  SELECT app_id, valid_until as updated FROM draft.apps_components_all
   UNION
  SELECT app_id, deployments_all.valid_from as updated FROM draft.apps_components_all JOIN draft.deployments_all on deployments_all.component_id = apps_components_all.component_id
   UNION
  SELECT app_id, deployments_all.valid_until as updated FROM draft.apps_components_all JOIN draft.deployments_all on deployments_all.component_id = apps_components_all.component_id
) SELECT updated FROM apps_history WHERE app_id = 'demo' AND updated IS NOT NULL ORDER BY updated;

WITH history AS (
  SELECT apps.name as app_name,
         components.name as component_name, apps_components.valid_from as component_valid_from, apps_components.valid_until as component_valid_until,
         image_ref as image, deployments.valid_from as deployment_valid_from, deployments.valid_until as deployment_valid_until
    FROM draft.apps
   INNER JOIN draft.apps_components_all apps_components
      ON apps_components.app_id = apps.id
   INNER JOIN draft.components
      ON components.id = apps_components.component_id
    LEFT JOIN draft.deployments_all deployments
      ON deployments.component_id = components.id
     AND deployments.valid_from <= apps_components.valid_from
     AND (apps_components.valid_until IS NULL OR deployments.valid_until IS NULL OR deployments.valid_until >= apps_components.valid_until)
   WHERE apps.id = 'demo'
   ORDER BY component_valid_from, deployment_valid_from
) SELECT * FROM history;

SELECT *
  FROM draft.apps
 INNER JOIN draft.apps_components_history
    ON apps_components_history.app_id = apps.id
 INNER JOIN draft.components
    ON components.id = apps_components_history.component_id
  LEFT JOIN draft.deployments_history
    ON deployments_history.component_id = components.id
 WHERE apps.id = 'demo'
;
