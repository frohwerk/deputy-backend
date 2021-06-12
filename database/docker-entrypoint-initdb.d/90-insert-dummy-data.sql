-- INSERT INTO apps VALUES
--   ('4b1f4ba7-ea9c-4042-b175-2ef3a60629e8', 'Example Application #1'),
--   ('6e07f5bd-4f30-4fa9-b981-ea4324c66be9', 'Example Application #2'),
--   ('4337db1f-cda0-44b0-ad6a-3c749b2dcef2', 'Example Application #3'),
--   ('bea8ba7d-45f4-45b2-9ff7-164a8095e065', 'Example Application #4');

-- INSERT INTO components (id, name, image) VALUES
--   ('5bcfb531-f113-436b-8c28-81fd9f365856', 'Example Component #1', '172.30.1.1:5000/myproject/example:1.0'),
--   ('db56fe67-d013-444b-9022-d0d10117e216', 'Example Component #2', '172.30.1.1:5000/myproject/example:1.0'),
--   ('c86a651c-893f-4b8c-afec-ed1466594ba3', 'Example Component #3', '172.30.1.1:5000/myproject/example:1.0'),
--   ('9766788d-8311-4754-97cf-35d1fcc1a973', 'Example Component #4', '172.30.1.1:5000/myproject/example:1.0'),
--   ('2c1e4c3a-6902-478a-992b-3c5c5006148b', 'Example Component #5', '172.30.1.1:5000/myproject/example:1.0'),
--   ('041966cd-f0a9-48d9-bbe2-976962d7e49c', 'Example Component #6', '172.30.1.1:5000/myproject/example:1.0');

-- INSERT INTO apps_components VALUES
--   ('4b1f4ba7-ea9c-4042-b175-2ef3a60629e8','5bcfb531-f113-436b-8c28-81fd9f365856'),
--   ('4b1f4ba7-ea9c-4042-b175-2ef3a60629e8','db56fe67-d013-444b-9022-d0d10117e216'),
--   ('4b1f4ba7-ea9c-4042-b175-2ef3a60629e8','c86a651c-893f-4b8c-afec-ed1466594ba3'),
--   ('6e07f5bd-4f30-4fa9-b981-ea4324c66be9','9766788d-8311-4754-97cf-35d1fcc1a973'),
--   ('6e07f5bd-4f30-4fa9-b981-ea4324c66be9','2c1e4c3a-6902-478a-992b-3c5c5006148b');

INSERT INTO envs (id, name) VALUES
  ('e7ccea48-c007-4ff5-b2fb-74516e77da00', 'Test'),
  ('c8f1d2d6-8305-48d6-a613-23cdb67b5a19', 'Produktion')
  ;

INSERT INTO platforms (id, env_id, name, api_server, namespace, secret) VALUES
  ('c49ca75c-da18-4641-950c-f5609877828f', 'e7ccea48-c007-4ff5-b2fb-74516e77da00', 'minishift / myproject', 'https://192.168.178.31:8443', 'myproject', 'eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJteXByb2plY3QiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlY3JldC5uYW1lIjoiZGVwdXR5LXRva2VuLXJncGh6Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6ImRlcHV0eSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjgwMTFiMDk2LWFjYTktMTFlYi05YjE0LTAwMTU1ZDYzMDEwOCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpteXByb2plY3Q6ZGVwdXR5In0.VRdGoGmkesFga1GU0ooP2KbwSzuq5zb9c3mNc8j0KGYd-eFe1-39FAG4TJU2is1b0tble5SF3TB0e4x4xFlBNNEtV2jUm7htOm0le0av6KtdTaGJA3WYhLKg_BD5G8Xq9irjRZg_rp448g1Bw03yzjF-YuOeWc9T95LMcT4bGarun1QxAPAx2ZBRNZxOZe7640x1X2s3qW5XocOSRRsBmtkpC-nJ-QYvlZsRGheU8-XSGT-gy-jDKU3KFOTA4dDsZSLgkmYzK4tb1hQEYKnUbH2Jjd74dIKpgMT27a_N77TS1-b36KGltaZEBEt7kfcHXHKijXrMCzJLEHOPOCEvXw'),
  ('3146c2ee-bdd7-40ed-83c2-fe9efdff4a95', 'c8f1d2d6-8305-48d6-a613-23cdb67b5a19', 'minishift / demo-prod', 'https://192.168.178.31:8443', 'demo-prod', 'eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZW1vLXByb2QiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlY3JldC5uYW1lIjoiZGVwdXR5LXRva2VuLXhzN3RxIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6ImRlcHV0eSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6ImNiMmI2MDI3LWIwOTctMTFlYi05YjE0LTAwMTU1ZDYzMDEwOCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZW1vLXByb2Q6ZGVwdXR5In0.GBOBM_laT3hEaQNiGTwdv5Vi17fO-hrkNqHfhcjYT91FPHd836S71_nh1L4cN536ZgFG4rNC11wvTWXN5l012Rv18T7MSkP_0IfuF2HKKmhnd2g0Pnl0b4yUFRH_WUx3gwyJIj2eyaVuf6wdS_zYwW-LmvXZeZcbPwQsX3yusFc6CgSrvpNOBPiWa8VlGA6fwVprsZLELCJ73W39WO6M9iGdab5RjmyichR-alfH-CpUQDRu7L2TlnL_7CUvy724yj5xAPh52oNjxVpvXUvACfBgNxVnmTMZDOK9EzwdRmvgIMTv4nm2Kc5SPQsvxA3c7F3K2PIbVFCCP7Na3KGtjA')
  ;
