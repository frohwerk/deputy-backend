SELECT gen_random_uuid();
SELECT * FROM apps;
SELECT * FROM components;
SELECT * FROM apps_components;

SELECT a.name, c.name, c.image FROM apps a JOIN apps_components r ON r.app_id = a.id LEFT JOIN components c ON c.id = r.component_id WHERE a.name = 'Example Application #1';

SELECT * FROM components;
SELECT * FROM apps_components JOIN components on components.id = apps_components.component_id;

SELECT * FROM components c WHERE NOT EXISTS (SELECT * FROM apps_components ac WHERE ac.component_id = c.id);
