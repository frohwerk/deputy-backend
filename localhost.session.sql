SELECT gen_random_uuid();
SELECT * FROM apps;
SELECT * FROM components;
SELECT * FROM apps_components;

SELECT * FROM components c WHERE NOT EXISTS (SELECT * FROM apps_components ac WHERE ac.component_id = c.id);

DELETE FROM apps WHERE name = 'Muffins';

UPDATE apps SET name = 'Banana Pie' WHERE id = '6e07f5bd-4f30-4fa9-b981-ea4324c66be9';
