SELECT routine_catalog, routine_schema, routine_name, routine_type, external_language, routine_definition FROM information_schema.routines WHERE routine_schema = 'public';
SELECT trigger_name, event_manipulation, event_object_catalog, event_object_schema, event_object_table, action_timing, action_statement FROM information_schema.triggers WHERE trigger_schema = 'public';

DROP SCHEMA IF EXISTS draft;
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
    app_id   VARCHAR(36) NOT NULL,
    component_id VARCHAR(36) NOT NULL,
    updated  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (app_id, component_id),
    FOREIGN KEY (app_id) REFERENCES draft.apps (id) ON DELETE CASCADE,
    FOREIGN KEY (component_id) REFERENCES draft.components (id) ON DELETE CASCADE
);

CREATE TABLE draft.apps_components_history (
    app_id        VARCHAR(36) NOT NULL,
    component_id  VARCHAR(36) NOT NULL,
    valid_from    TIMESTAMP NOT NULL,
    valid_until   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (app_id) REFERENCES draft.apps (id) ON DELETE CASCADE,
    FOREIGN KEY (component_id) REFERENCES draft.components (id) ON DELETE CASCADE,
    PRIMARY KEY (app_id, component_id, valid_from)
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

------------------------------------------------------------------------------------------------------------------------------------------------
SELECT * FROM deployments WHERE char_length(platform_id) < 20 OR char_length(component_id) < 20;
SELECT * FROM apps_components WHERE char_length(app_id) < 20 OR char_length(component_id) < 20;
SELECT * FROM deployments_history WHERE char_length(platform_id) < 20 OR char_length(component_id) < 20;
SELECT * FROM apps_components_history WHERE char_length(app_id) < 20 OR char_length(component_id) < 20;

SELECT * FROM apps_timeline WHERE char_length(app_id) < 20;
------------------------------------------------------------------------------------------------------------------------------------------------
SELECT * FROM apps_history WHERE app_id = 'tester' AND env_id IN ('-example', 'integration') ORDER BY 1, 2, 3, 4;
SELECT * FROM apps_history WHERE app_id = 'tester' AND env_id IN ('example', '-integration') ORDER BY 1, 2, 3, 4;

SELECT * FROM apps_timeline WHERE app_id = '0f99b1fd-b546-4eb3-9977-a05724257d53' AND env_id = 'e7ccea48-c007-4ff5-b2fb-74516e77da00';
SELECT * FROM apps_components_all WHERE app_id = '0f99b1fd-b546-4eb3-9977-a05724257d53';
SELECT * FROM deployments_all WHERE platform_id = 'c49ca75c-da18-4641-950c-f5609877828f' AND component_id IN ('4baf7782-35ea-44c0-a6a5-05724f001fa2', '151d898a-f78b-41ba-8fca-4f5f1fb60bd4');
SELECT * FROM apps_history WHERE app_id = '0f99b1fd-b546-4eb3-9977-a05724257d53' AND env_id = 'e7ccea48-c007-4ff5-b2fb-74516e77da00' ORDER BY 1, 2, 3, 4;

SELECT * FROM apps_timeline WHERE app_id = '0f99b1fd-b546-4eb3-9977-a05724257d53' AND env_id = 'e7ccea48-c007-4ff5-b2fb-74516e77da00';
SELECT * FROM apps_components WHERE app_id = '0f99b1fd-b546-4eb3-9977-a05724257d53';
SELECT * FROM apps_components_history WHERE app_id = '0f99b1fd-b546-4eb3-9977-a05724257d53';
SELECT * FROM deployments WHERE platform_id = 'c49ca75c-da18-4641-950c-f5609877828f';
SELECT * FROM deployments_history WHERE platform_id = 'c49ca75c-da18-4641-950c-f5609877828f';
------------------------------------------------------------------------------------------------------------------------------------------------
SELECT * FROM pg_timezone_names;
SELECT ROUND(EXTRACT(EPOCH FROM NULL::TIMESTAMP AT TIME ZONE 'Europe/Berlin')) now;
SELECT '2021-06-14T13:28:32.681101Z'::TIMESTAMP AT TIME ZONE 'UTC' TS, EXTRACT(EPOCH FROM '2021-06-14T13:28:32.681101Z' AT TIME ZONE 'UTC') EPOCH;
SELECT CURRENT_TIMESTAMP AT TIME ZONE 'UTC', to_char(CURRENT_TIMESTAMP AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS.USZ') now;
SELECT to_timestamp(1623667768.448556), to_timestamp(1623667768.448556) AT TIME ZONE 'UTC';
SELECT to_timestamp(1623667768) AT TIME ZONE 'Europe/Berlin';
SELECT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP AT TIME ZONE 'UTC') now;
SELECT to_char(NULL::TIMESTAMP, 'YYYY-MM-DD HH24:MI:SS.MS') now;
------------------------------------------------------------------------------------------------------------------------------------------------
SELECT app_id, env_id, valid_from FROM apps_timeline
 WHERE app_id = '0f99b1fd-b546-4eb3-9977-a05724257d53' AND env_id = 'e7ccea48-c007-4ff5-b2fb-74516e77da00'
 ORDER BY valid_from DESC;
WITH slice AS (
SELECT app_id, env_id, valid_from FROM apps_timeline
 WHERE app_id = '0f99b1fd-b546-4eb3-9977-a05724257d53' AND env_id = 'e7ccea48-c007-4ff5-b2fb-74516e77da00' AND valid_from < '2021-06-14 10:33:44.448556'
 ORDER BY valid_from DESC FETCH FIRST ROW ONLY
) SELECT * FROM slice;
------------------------------------------------------------------------------------------------------------------------------------------------
SELECT app_id, env_id, valid_from
  FROM apps_timeline
 WHERE app_id = '0f99b1fd-b546-4eb3-9977-a05724257d53'
   AND env_id = 'e7ccea48-c007-4ff5-b2fb-74516e77da00'
   AND valid_from < '2021-06-14 10:33:44.448556'
 ORDER BY valid_from DESC FETCH FIRST ROW ONLY;
------------------------------------------------------------------------------------------------------------------------------------------------
WITH slice AS (
  SELECT app_id, env_id, valid_from, '2021-06-14 10:33:44.448556'::TIMESTAMP FROM apps_timeline
   WHERE app_id = '0f99b1fd-b546-4eb3-9977-a05724257d53' AND env_id = 'e7ccea48-c007-4ff5-b2fb-74516e77da00' AND valid_from < '2021-06-14 10:33:44.448556'
   ORDER BY valid_from DESC FETCH FIRST ROW ONLY
) SELECT * FROM slice
------------------------------------------------------------------------------------------------------------------------------------------------
SELECT * FROM apps;
------------------------------------------------------------------------------------------------------------------------------------------------
SELECT MAX(apps_timeline.valid_from) AS valid_from FROM apps_timeline
 WHERE apps_timeline.app_id = '0f99b1fd-b546-4eb3-9977-a05724257d53' AND apps_timeline.env_id = 'e7ccea48-c007-4ff5-b2fb-74516e77da00' AND apps_timeline.valid_from < '2021-06-14 10:33:44.448556'::TIMESTAMP;
------------------------------------------------------------------------------------------------------------------------------------------------
SELECT EXTRACT(EPOCH FROM '2021-06-15T08:13:38.671053Z'::TIMESTAMP WITH TIME ZONE);
------------------------------------------------------------------------------------------------------------------------------------------------
SELECT * FROM apps_history;
DELETE FROM apps WHERE id = 'test';
INSERT INTO apps (id, name) VALUES('test', 'Test');
INSERT INTO apps_timeline (app_id, env_id, valid_from)
  SELECT 'd9420583-212c-4613-a451-a4c7b759e184', envs.id, '2021-06-15 08:11:38.671053' FROM apps CROSS JOIN envs
  WHERE apps.id = 'd9420583-212c-4613-a451-a4c7b759e184';
------------------------------------------------------------------------------------------------------------------------------------------------
WITH
params AS (
  SELECT '0f99b1fd-b546-4eb3-9977-a05724257d53' _app_id, 'e7ccea48-c007-4ff5-b2fb-74516e77da00' _env_id, to_timestamp(1623689711.0) AT TIME ZONE 'UTC' _timestamp
),
slice AS (
  SELECT
    (SELECT MAX(valid_from) FROM params, apps_timeline WHERE app_id = _app_id AND env_id = _env_id AND valid_from < _timestamp) AS valid_from,
    (SELECT MIN(valid_from) FROM params, apps_timeline WHERE app_id = _app_id AND env_id = _env_id AND valid_from >= _timestamp) AS valid_until
)
SELECT apps.id AS app_id, apps.name AS app_name,
        slice.valid_from, slice.valid_until,
        COALESCE(components.id, '') AS component_id, COALESCE(components.name, '') AS component_name,
        COALESCE(h.image_ref, '') AS image_ref, COALESCE(to_char(h.last_deployment, 'YYYY-MM-DD HH24:MI:SS.USZ'), '') AS last_deployment
FROM params CROSS JOIN slice
INNER JOIN apps_history h ON h.app_id = _app_id AND h.env_id = _env_id AND h.valid_from = slice.valid_from
INNER JOIN apps ON apps.id = h.app_id
LEFT JOIN components ON components.id = h.component_id
ORDER BY 3 DESC, 5 ASC;
------------------------------------------------------------------------------------------------------------------------------------------------
WITH
params AS (
  SELECT '0f99b1fd-b546-4eb3-9977-a05724257d53' _app_id, 'e7ccea48-c007-4ff5-b2fb-74516e77da00' _env_id, '2021-06-15 10:35:44.448556'::TIMESTAMP _timestamp
--  SELECT '0f99b1fd-b546-4eb3-9977-a05724257d53' _app_id, 'e7ccea48-c007-4ff5-b2fb-74516e77da00' _env_id, '2021-06-14 10:33:44.448556'::TIMESTAMP _timestamp
),
slice AS (
  SELECT
    (SELECT MAX(apps_timeline.valid_from) FROM params, apps_timeline WHERE app_id = _app_id AND env_id = _env_id AND valid_from < _timestamp) AS valid_from,
    (SELECT MIN(apps_timeline.valid_from) FROM params, apps_timeline WHERE app_id = _app_id AND env_id = _env_id AND valid_from >= _timestamp) AS valid_until
--) SELECT _app_id app_id, _env_id env_id, slice.* FROM params, slice;
)
SELECT apps.id AS app_id, apps.name AS app_name,
       slice.valid_from, slice.valid_until,
       COALESCE(components.id, '') AS component_id, COALESCE(components.name, '') AS component_name,
       COALESCE(apps_history.image_ref, '') AS image, COALESCE(to_char(deployed, 'YYYY-MM-DD HH24:MI:SS.USZ'), '') AS last_deployment
  FROM params CROSS JOIN slice
  JOIN apps ON apps.id = _app_id
  JOIN apps_history ON apps_history.app_id = apps.id AND apps_history.env_id = _env_id
  JOIN components ON components.id = apps_history.component_id
 WHERE apps_history.valid_from = slice.valid_from
 ORDER BY 5;
------------------------------------------------------------------------------------------------------------------------------------------------
SELECT * FROM apps_history WHERE app_id = '0f99b1fd-b546-4eb3-9977-a05724257d53' AND env_id = 'e7ccea48-c007-4ff5-b2fb-74516e77da00' ORDER BY valid_from DESC, component_id ASC;
WITH
params AS (
--  SELECT '0f99b1fd-b546-4eb3-9977-a05724257d53' _app_id, 'e7ccea48-c007-4ff5-b2fb-74516e77da00' _env_id, '2022-06-15 10:33:44.448556'::TIMESTAMP _timestamp
  SELECT '0f99b1fd-b546-4eb3-9977-a05724257d53' _app_id, 'e7ccea48-c007-4ff5-b2fb-74516e77da00' _env_id, '2021-06-14 13:28:32.681101Z'::TIMESTAMP _timestamp
),
slice AS (
  SELECT
    (SELECT MAX(valid_from) FROM params, apps_timeline WHERE app_id = _app_id AND env_id = _env_id AND valid_from < _timestamp) AS valid_from,
    (SELECT MIN(valid_from) FROM params, apps_timeline WHERE app_id = _app_id AND env_id = _env_id AND valid_from >= _timestamp) AS valid_until
) SELECT * FROM slice;
------------------------------------------------------------------------------------------------------------------------------------------------
SELECT EXTRACT(EPOCH FROM '2021-06-14 13:28:32.681101Z'::TIMESTAMP WITH TIME ZONE);
------------------------------------------------------------------------------------------------------------------------------------------------
-- Lösung: Lesen eines bestimmten Standes aus der apps_history
WITH
params AS (
  SELECT '0f99b1fd-b546-4eb3-9977-a05724257d53' _app_id,
         'e7ccea48-c007-4ff5-b2fb-74516e77da00' _env_id,
         to_timestamp(1623677312.681101) AT TIME ZONE 'UTC' _timestamp
),
slice AS (
  SELECT
    (SELECT MAX(valid_from) FROM params, apps_timeline WHERE app_id = _app_id AND env_id = _env_id AND valid_from < _timestamp) AS valid_from,
    (SELECT MIN(valid_from) FROM params, apps_timeline WHERE app_id = _app_id AND env_id = _env_id AND valid_from >= _timestamp) AS valid_until
)
SELECT apps.id AS app_id, apps.name AS app_name,
       slice.valid_from, slice.valid_until,
       COALESCE(components.id, '') AS component_id, COALESCE(components.name, '') AS component_name,
       COALESCE(h.image_ref, '') AS image_ref, COALESCE(to_char(h.last_deployment, 'YYYY-MM-DD HH24:MI:SS.USZ'), '') AS last_deployment
  FROM params CROSS JOIN slice
 INNER JOIN apps_history h ON h.app_id = _app_id AND h.env_id = _env_id AND h.valid_from = slice.valid_from
 INNER JOIN apps ON apps.id = h.app_id
  LEFT JOIN components ON components.id = h.component_id
 ORDER BY 3 DESC, 5 ASC;
------------------------------------------------------------------------------------------------------------------------------------------------
--)
SELECT apps.id AS app_id, apps.name AS app_name, slice.valid_from, slice.valid_until,
       COALESCE(components.id, '') AS component_id, COALESCE(components.name, '') AS component_name,
       COALESCE(apps_history.image_ref, '') AS image, COALESCE(apps_history.deployed, '0001-01-01 00:00:00'::TIMESTAMP) AS deployed
  FROM params, slice, apps_history
  JOIN apps ON apps.id = slice.app_id
  JOIN envs ON envs.id = slice.env_id
  LEFT JOIN components ON components.id = component_id
 WHERE apps_history.app_id = _app_id AND apps_history.env_id = _env_id AND apps_history.valid_from = slice.valid_from
 ORDER BY 5;
------------------------------------------------------------------------------------------------------------------------------------------------
SELECT apps.id, apps.name, valid_from,
       ROW_NUMBER() OVER (ORDER BY 3 DESC, 5) - 1 AS age,
       COALESCE(components.id, '') cid, COALESCE(components.name, '') cname,
       COALESCE(image_ref, '') cimage, COALESCE(deployed, '0001-01-01 00:00:00'::TIMESTAMP) cdeployed
  FROM apps_history
  JOIN apps ON apps.id = app_id
  JOIN envs ON envs.id = env_id
  LEFT JOIN components ON components.id = component_id
 WHERE app_id = '0f99b1fd-b546-4eb3-9977-a05724257d53' AND env_id = 'e7ccea48-c007-4ff5-b2fb-74516e77da00'
 ORDER BY 3 DESC, 5;
------------------------------------------------------------------------------------------------------------------------------------------------
INSERT INTO apps_timeline
SELECT app_id, env_id, '2021-06-14 07:39:39.402386'::TIMESTAMP FROM apps_timeline
ON CONFLICT DO NOTHING;
------------------------------------------------------------------------------------------------------------------------------------------------
SELECT * FROM apps_timeline WHERE app_id = '0f99b1fd-b546-4eb3-9977-a05724257d53' AND env_id = 'e7ccea48-c007-4ff5-b2fb-74516e77da00';
SELECT envs.name env_name, apps.name app_name, valid_from, apps_history.component_id, components.name component_name, image_ref, deployed
  FROM apps_history
  JOIN apps ON apps.id = app_id
  JOIN envs ON envs.id = env_id
  LEFT JOIN components ON components.id = component_id
 WHERE app_id = '0f99b1fd-b546-4eb3-9977-a05724257d53' AND env_id = 'e7ccea48-c007-4ff5-b2fb-74516e77da00' ORDER BY 1, 2, 3 DESC, 4;
------------------------------------------------------------------------------------------------------------------------------------------------
  SELECT t.app_id, p.env_id, t.valid_from, c.component_id, d.image_ref
    FROM platforms p
   CROSS JOIN apps_timeline t
   INNER JOIN apps_components_all c
      ON c.app_id = t.app_id
     AND c.valid_from <= t.valid_from AND t.valid_from < COALESCE(c.valid_until, CURRENT_TIMESTAMP)
    LEFT JOIN deployments_all d
      ON d.component_id = c.component_id AND d.platform_id = p.id
     AND d.valid_from <= t.valid_from AND t.valid_from < COALESCE(d.valid_until, CURRENT_TIMESTAMP)
 WHERE t.app_id = 'tester' AND p.env_id = 'integration'
 ORDER BY 1, 2, 3, 4;
------------------------------------------------------------------------------------------------------------------------------------------------
