INSERT INTO apps VALUES
  ('4b1f4ba7-ea9c-4042-b175-2ef3a60629e8', 'Example Application #1'),
  ('6e07f5bd-4f30-4fa9-b981-ea4324c66be9', 'Example Application #2'),
  ('4337db1f-cda0-44b0-ad6a-3c749b2dcef2', 'Example Application #3'),
  ('bea8ba7d-45f4-45b2-9ff7-164a8095e065', 'Example Application #4');

INSERT INTO components VALUES
  ('5bcfb531-f113-436b-8c28-81fd9f365856', 'Example Component #1', '172.30.1.1:5000/myproject/example:1.0'),
  ('db56fe67-d013-444b-9022-d0d10117e216', 'Example Component #2', '172.30.1.1:5000/myproject/example:1.0'),
  ('c86a651c-893f-4b8c-afec-ed1466594ba3', 'Example Component #3', '172.30.1.1:5000/myproject/example:1.0'),
  ('9766788d-8311-4754-97cf-35d1fcc1a973', 'Example Component #4', '172.30.1.1:5000/myproject/example:1.0'),
  ('2c1e4c3a-6902-478a-992b-3c5c5006148b', 'Example Component #5', '172.30.1.1:5000/myproject/example:1.0'),
  ('041966cd-f0a9-48d9-bbe2-976962d7e49c', 'Example Component #6', '172.30.1.1:5000/myproject/example:1.0');

INSERT INTO apps_components VALUES
  ('4b1f4ba7-ea9c-4042-b175-2ef3a60629e8','5bcfb531-f113-436b-8c28-81fd9f365856'),
  ('4b1f4ba7-ea9c-4042-b175-2ef3a60629e8','db56fe67-d013-444b-9022-d0d10117e216'),
  ('4b1f4ba7-ea9c-4042-b175-2ef3a60629e8','c86a651c-893f-4b8c-afec-ed1466594ba3'),
  ('6e07f5bd-4f30-4fa9-b981-ea4324c66be9','9766788d-8311-4754-97cf-35d1fcc1a973'),
  ('6e07f5bd-4f30-4fa9-b981-ea4324c66be9','2c1e4c3a-6902-478a-992b-3c5c5006148b');
