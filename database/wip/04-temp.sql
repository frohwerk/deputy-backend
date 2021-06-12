SELECT * FROM envs WHERE char_length(id) < 20;
SELECT * FROM platforms WHERE char_length(id) < 20;
SELECT * FROM apps WHERE char_length(id) < 20;
SELECT * FROM components WHERE char_length(id) < 20;
SELECT * FROM deployments WHERE char_length(platform_id) < 20 OR char_length(component_id) < 20;
SELECT * FROM apps_components WHERE char_length(app_id) < 20 OR char_length(component_id) < 20;

DELETE FROM envs WHERE char_length(id) < 20;
DELETE FROM platforms WHERE char_length(id) < 20;
DELETE FROM apps WHERE char_length(id) < 20;
DELETE FROM components WHERE char_length(id) < 20;
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

DELETE FROM apps_components WHERE app_id = 'tester';
DELETE FROM deployments WHERE char_length(platform_id) < 20 OR char_length(component_id) < 20;
DELETE FROM deployments_history WHERE char_length(platform_id) < 20 OR char_length(component_id) < 20;
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
---------------------------------------------------------------------------------------------------------------------------------------------------
UPDATE deployments SET image_ref = 'image-registry.cluster.local/my-namespace/b:5.0.1.RELEASE' WHERE component_id = 'component-b' AND platform_id = 'minishift-si';
