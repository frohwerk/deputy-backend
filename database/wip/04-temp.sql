SELECT * FROM deployments_all;
SELECT * FROM apps_components_all;
---------------------------------------------------------------------------------------------------------------------------------------------------
SELECT * FROM envs WHERE char_length(id) < 20;
SELECT * FROM platforms WHERE char_length(id) < 20;
SELECT * FROM apps WHERE char_length(id) < 20;
SELECT * FROM components WHERE char_length(id) < 20;
SELECT * FROM deployments WHERE char_length(platform_id) < 20 OR char_length(component_id) < 20;
SELECT * FROM apps_components WHERE char_length(app_id) < 20 OR char_length(component_id) < 20;
SELECT * FROM deployments_history WHERE char_length(platform_id) < 20 OR char_length(component_id) < 20;
SELECT * FROM apps_components_history WHERE char_length(app_id) < 20 OR char_length(component_id) < 20;

DELETE FROM envs WHERE char_length(id) < 20;
DELETE FROM apps WHERE char_length(id) < 20;
DELETE FROM components WHERE char_length(id) < 20;

DELETE FROM platforms WHERE char_length(id) < 20;
DELETE FROM deployments WHERE char_length(platform_id) < 20 OR char_length(component_id) < 20;
DELETE FROM apps_components WHERE char_length(app_id) < 20 OR char_length(component_id) < 20;

---------------------------------------------------------------------------------------------------------------------------------------------------
-- scenario #1

INSERT INTO envs (id, name) VALUES ('example', 'Example');
INSERT INTO platforms (id, env_id, name, api_server, namespace, secret) VALUES ('minishift', 'example', 'Minishift', 'https://192.168.178.31:8443', 'my-namespace', '');
INSERT INTO apps (id, name) VALUES ('tester', 'Test-Anwendung');
INSERT INTO components (id, name) VALUES ('component-a', 'Irgendeine Komponente');
INSERT INTO components (id, name) VALUES ('component-b', 'Eine andere Komponente');
INSERT INTO deployments (component_id, platform_id, image_ref) VALUES ('component-a', 'minishift', 'image-registry.cluster.local/my-namespace/a:1.0.2');
INSERT INTO deployments (component_id, platform_id, image_ref) VALUES ('component-b', 'minishift', 'image-registry.cluster.local/my-namespace/b:4.1');

---------------------------------------------------------------------------------------------------------------------------------------------------
-- scenario #2 (bulit on top of scenario #1)

INSERT INTO envs (id, name) VALUES ('integration', 'Environment for user acceptance testing');
INSERT INTO platforms (id, env_id, name, api_server, namespace, secret) VALUES ('minishift-si', 'integration', 'Minishift (SI)', 'https://192.168.178.31:8443', 'my-namespace', '');
INSERT INTO deployments (component_id, platform_id, image_ref) VALUES ('component-a', 'minishift-si', 'image-registry.cluster.local/my-namespace/a:1.0.2');

---------------------------------------------------------------------------------------------------------------------------------------------------

DELETE FROM deployments_history WHERE char_length(platform_id) < 20 OR char_length(component_id) < 20;
DELETE FROM deployments WHERE char_length(platform_id) < 20 OR char_length(component_id) < 20;
DELETE FROM deployments_history WHERE char_length(platform_id) < 20 OR char_length(component_id) < 20;
DELETE FROM apps_components_history WHERE char_length(app_id) < 20 OR char_length(component_id) < 20;
DELETE FROM apps_components WHERE app_id = 'tester';
DELETE FROM apps_components_history WHERE char_length(app_id) < 20 OR char_length(component_id) < 20;
DELETE FROM apps_timeline WHERE char_length(app_id) < 20;

---------------------------------------------------------------------------------------------------------------------------------------------------

SELECT * FROM deployments WHERE char_length(platform_id) < 20 OR char_length(component_id) < 20;
SELECT * FROM apps_components WHERE char_length(app_id) < 20 OR char_length(component_id) < 20;
SELECT * FROM deployments_history WHERE char_length(platform_id) < 20 OR char_length(component_id) < 20;
SELECT * FROM apps_components_history WHERE char_length(app_id) < 20 OR char_length(component_id) < 20;
SELECT * FROM apps_timeline WHERE char_length(app_id) < 20;

SELECT * FROM apps_history WHERE app_id = 'tester' order by 1, 2, 3, 4;

-- test case #1
INSERT INTO apps_components (app_id, component_id) VALUES ('tester', 'component-a');
-- test case #2
INSERT INTO apps_components (app_id, component_id) VALUES ('tester', 'component-b');
-- test case #3
DELETE FROM apps_components WHERE app_id = 'tester' AND component_id = 'component-b';
-- test case #4
INSERT INTO apps_components (app_id, component_id) VALUES ('tester', 'component-b');
-- test case #5
INSERT INTO deployments (component_id, platform_id, image_ref) VALUES ('component-b', 'minishift-si', 'image-registry.cluster.local/my-namespace/a:1.0.3');
-- test case #6
UPDATE deployments SET image_ref = 'image-registry.cluster.local/my-namespace/b:4.1' WHERE component_id = 'component-b' AND platform_id = 'minishift-si';
-- test case #7
DELETE FROM deployments WHERE component_id = 'component-a' AND platform_id = 'minishift';
---------------------------------------------------------------------------------------------------------------------------------------------------
UPDATE deployments SET image_ref = 'image-registry.cluster.local/my-namespace/b:5.1.0' WHERE component_id = 'component-b' AND platform_id = 'minishift-si';
---------------------------------------------------------------------------------------------------------------------------------------------------
SELECT app_id, env_id, '2021-06-03 15:59:17.156'::timestamp FROM platforms CROSS JOIN apps_components
 WHERE platforms.id = 'minishift-si' AND apps_components.component_id = 'component-b';
---------------------------------------------------------------------------------------------------------------------------------------------------
SELECT * FROM apps_timeline;
SELECT t.app_id, t.env_id, t.valid_from, c.component_id, d.image_ref
  FROM apps_timeline t
 INNER JOIN platforms p
    ON p.env_id = t.env_id
 INNER JOIN apps_components_all c
    ON c.app_id = t.app_id
   AND c.valid_from <= t.valid_from AND t.valid_from < COALESCE(c.valid_until, CURRENT_TIMESTAMP)
  LEFT JOIN deployments_all d
    ON d.component_id = c.component_id AND d.platform_id = p.id
   AND d.valid_from <= t.valid_from AND t.valid_from < COALESCE(d.valid_until, CURRENT_TIMESTAMP)
 WHERE t.app_id = 'tester' AND p.env_id IN ('example') ORDER BY 1, 2, 3, 4;
---------------------------------------------------------------------------------------------------------------------------------------------------
DROP TABLE _experiment;
CREATE TABLE _experiment (
  app_id VARCHAR(36) NOT NULL,
  env_id VARCHAR(36) NOT NULL,
  PRIMARY KEY (app_id, env_id)
);

DELETE FROM _experiment;
SELECT * FROM _experiment;
INSERT INTO _experiment (app_id, env_id)
SELECT 'application-a', envs.id FROM envs WHERE char_length(envs.id) < 20
ON CONFLICT DO NOTHING;
