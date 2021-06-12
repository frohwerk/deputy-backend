SELECT routine_catalog, routine_schema, routine_name, routine_type, external_language, routine_definition FROM information_schema.routines WHERE routine_schema = 'public';
SELECT trigger_name, event_manipulation, event_object_catalog, event_object_schema, event_object_table, action_timing, action_statement FROM information_schema.triggers WHERE trigger_schema = 'public';

SELECT * FROM envs;
SELECT * FROM platforms;
SELECT * FROM apps;
SELECT * FROM components;
SELECT * FROM deployments;

SELECT * FROM deployments_history;

SELECT * FROM apps_components;

SELECT * FROM files;

INSERT INTO apps(id, name) VALUES ('555d8b8f-0eed-4a6c-a8a1-ca16f579aef2', 'Demo');
INSERT INTO apps_components(app_id, component_id) VALUES('555d8b8f-0eed-4a6c-a8a1-ca16f579aef2', 'b707cbc2-967d-4703-8db4-7feb24a71360');

SELECT e.env_name as env, a.name as app, c.name as comp, d.image_ref, d.updated
  FROM apps a
  JOIN envs e ON 1 = 1
  JOIN apps_components ac ON ac.app_id = a.id
  JOIN components c ON c.component_id = ac.component_id
  JOIN platforms p ON p.pf_env = e.env_id
  JOIN deployments d ON d.component_id = c.component_id AND d.platform_id = p.pf_id
WHERE a.id = '555d8b8f-0eed-4a6c-a8a1-ca16f579aef2' AND e.env_id = 'e7ccea48-c007-4ff5-b2fb-74516e77da00';

CREATE TABLE deployments_history (
    component_id VARCHAR(36) NOT NULL,
    platform_id  VARCHAR(36) NOT NULL,
    valid_from   TIMESTAMP NOT NULL,
    valid_until  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    image_ref    VARCHAR(150) NOT NULL,
    PRIMARY KEY (component_id, platform_id, valid_from),
    FOREIGN KEY (platform_id) REFERENCES platforms (pf_id) ON DELETE CASCADE,
    FOREIGN KEY (component_id) REFERENCES components (component_id) ON DELETE CASCADE
);

DELETE FROM deployments_history WHERE component_id = '9ff53c7a-1451-424d-b262-5ae0b6f3c65b' AND platform_id = '3146c2ee-bdd7-40ed-83c2-fe9efdff4a95';

SELECT * FROM deployments WHERE component_id = '9ff53c7a-1451-424d-b262-5ae0b6f3c65b' AND platform_id = '3146c2ee-bdd7-40ed-83c2-fe9efdff4a95';
SELECT * FROM deployments_history WHERE component_id = '9ff53c7a-1451-424d-b262-5ae0b6f3c65b' AND platform_id = '3146c2ee-bdd7-40ed-83c2-fe9efdff4a95';

SELECT * FROM deployments;
SELECT * FROM deployments_history;

SELECT pg_notify('demo_channel', CONCAT('Hallo', ' ', 'Welt', '!'));

NOTIFY deployments, 'Test';

UPDATE deployments
SET image_ref = '172.30.1.1:5000/myproject/node-hello-world:1.0.4'
WHERE component_id = '9ff53c7a-1451-424d-b262-5ae0b6f3c65b' AND platform_id = '3146c2ee-bdd7-40ed-83c2-fe9efdff4a95';

DROP TRIGGER deployments_trigger_update ON deployments;
DROP FUNCTION deployment_history_update;
DROP PROCEDURE deployment_history_update;

CREATE OR REPLACE FUNCTION deployment_history_update() RETURNS trigger AS $$
  BEGIN
    NEW.updated := CURRENT_TIMESTAMP;
    INSERT INTO deployments_history (component_id, platform_id, valid_from, valid_until, image_ref)
    VALUES(OLD.component_id, OLD.platform_id, OLD.updated, NEW.updated, OLD.image_ref);
  END;
$$ LANGUAGE plpgsql;

--REFERENCING OLD TABLE AS old
CREATE TRIGGER deployments_trigger_update
BEFORE UPDATE ON deployments
FOR EACH ROW
EXECUTE PROCEDURE deployment_history_update();

SELECT *
  FROM apps_components ac
  JOIN platforms p ON 1 = 1
  JOIN components c ON c.component_id = ac.component_id
 WHERE ac.app_id = 'a3882005-2f7c-43a0-85f7-3a0375cec6b4' AND p.pf_env = 'e7ccea48-c007-4ff5-b2fb-74516e77da00';

SELECT c.name, COALESCE(d.image_ref, ''), d.updated
  FROM apps_components ac
  JOIN platforms p ON 1 = 1
  JOIN components c ON c.component_id = ac.component_id
  LEFT JOIN deployments d ON d.component_id = c.component_id AND d.platform_id = p.pf_id
 WHERE ac.app_id = 'a3882005-2f7c-43a0-85f7-3a0375cec6b4' AND p.pf_env = 'e7ccea48-c007-4ff5-b2fb-74516e77da00';

SELECT c.name, COALESCE(d.image_ref, ''), d.updated
  FROM apps_components ac
  JOIN platforms p ON 1 = 1
  JOIN components c ON c.component_id = ac.component_id
  LEFT JOIN deployments d ON d.component_id = c.component_id AND d.platform_id = p.pf_id
 WHERE ac.app_id = 'aca40db4-cea5-4a1d-bb87-271419dc51b5'
   AND p.pf_env = 'c8f1d2d6-8305-48d6-a613-23cdb67b5a19';

SELECT component_id, name
  FROM components c
 WHERE NOT EXISTS (SELECT * FROM apps_components ac WHERE ac.component_id = c.component_id)

SELECT c.component_id, c.name
  FROM components c
 WHERE NOT EXISTS (SELECT * FROM apps_components r WHERE r.component_id = c.component_id and r.app_id = '488b61ca-afb0-4e24-80c0-c2f2b212eee4')

DELETE FROM envs WHERE env_name in ('Produktion', 'Reisepass');
DELETE FROM platforms;

SELECT env_id, env_name FROM envs WHERE env_id = 'f12dfb6a-8abe-4aea-bbca-ba9ba47ac441';
SELECT * FROM platforms WHERE pf_id = '73015cfb-2605-46ce-9690-2f9dcc7d6595';
SELECT * FROM components c WHERE c.name not like 'Examp%';

SELECT c.id, c.name, c.updated, c.version, c.image, count(r.app_id) FROM components c LEFT JOIN apps_components r on r.component_id = c.id and r.app_id != '018cb51f-7bd9-4bbb-a6e7-31ef736b1d2c' GROUP BY c.id;
SELECT c.id, c.name, c.updated, c.image FROM components c WHERE NOT EXISTS (SELECT * FROM apps_components r WHERE r.component_id = c.id);

SELECT e.env_name, p.pf_name, p.pf_api_server, p.pf_namespace, c.name, c.updated, c.image
  FROM components c
  LEFT JOIN platforms_components r ON r.component_id = c.id
  JOIN platforms p ON p.pf_id = r.platform_id
  JOIN envs e ON e.env_id = p.pf_env;

SELECT a.name, c.name, c.image FROM apps a LEFT JOIN apps_components r ON r.app_id = a.id LEFT JOIN components c ON c.id = r.component_id where a.name not like 'Examp%';

SELECT a.name, c.name, c.image FROM apps a JOIN apps_components r ON r.app_id = a.id LEFT JOIN components c ON c.id = r.component_id WHERE a.id = 'a044fd27-d3e9-4eb8-b23a-8dd1c0af49e7';

SELECT *
  FROM apps
  JOIN apps_components ON apps.id = apps_components.app_id
  JOIN components ON components.id = apps_components.component_id
 WHERE apps.name not like 'Example Application #%';

SELECT * FROM components c WHERE name not like 'Example%';
SELECT * FROM components c WHERE NOT EXISTS (SELECT * FROM apps_components ac WHERE ac.component_id = c.id);

DELETE FROM components WHERE name not like 'Example%';

DELETE FROM files;

SELECT * FROM files;
DELETE FROM images_artifacts WHERE image_id = '172.30.1.1:5000/myproject/node-hello-world:1.0.3';
SELECT image_id, file_path, file_digest FROM images_artifacts JOIN files ON files.file_id = images_artifacts.file_id;
DELETE FROM files WHERE file_parent = 'e3387ae8-aa20-47bb-9b37-83966425628f' AND file_path like 'META-INF%';
SELECT * FROM components c WHERE NOT EXISTS (SELECT * FROM apps_components ac WHERE ac.component_id = c.id);
