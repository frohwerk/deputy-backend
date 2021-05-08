SELECT gen_random_uuid();
SELECT * FROM apps;
SELECT * FROM envs;
SELECT * FROM platforms;
SELECT * FROM components;
SELECT * FROM deployments;

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
