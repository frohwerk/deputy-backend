SELECT * FROM components;

MERGE INTO components
USING SELECT name FROM components WHERE name = 'Banane'
WHEN NOT MATCHED THEN INSERT (name) VALUES('Banane');

MERGE INTO components
USING SELECT image, version FROM components WHERE name = $1
WHEN MATCHED THEN UPDATE SET image = $2, version = $3
WHEN NOT MATCHED THEN INSERT (name, image, version) VALUES($1, $2, $3)
