-----------------------------------------------------------------------------------------------------------------------------------------------------------------
SELECT * FROM files;
-----------------------------------------------------------------------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION messages_notify() RETURNS trigger AS $$
  BEGIN
    CASE TG_OP
      WHEN 'INSERT' THEN
        PERFORM pg_notify(TG_TABLE_NAME, FORMAT("%s;%s", SUBSTR(TG_OP, 0, 1), NEW.ID));
        RETURN NEW;
      WHEN 'UPDATE' THEN
        PERFORM pg_notify(TG_TABLE_NAME, FORMAT("%s;%s", SUBSTR(TG_OP, 0, 1), NEW.ID));
        RETURN NEW;
      WHEN 'DELETE' THEN
        PERFORM pg_notify(TG_TABLE_NAME, FORMAT("%s;%s", SUBSTR(TG_OP, 0, 1), OLD.ID));
        RETURN OLD;
    END CASE;
    RETURN NEW;
  END;
$$ LANGUAGE plpgsql;
-----------------------------------------------------------------------------------------------------------------------------------------------------------------
CREATE TRIGGER messages_notify
AFTER INSERT OR UPDATE ON messages
FOR EACH ROW
EXECUTE FUNCTION messages_notify();
-----------------------------------------------------------------------------------------------------------------------------------------------------------------
DROP FUNCTION IF EXISTS messages_notify();
DROP TRIGGER IF EXISTS messages_notify ON messages;
-----------------------------------------------------------------------------------------------------------------------------------------------------------------
DELETE FROM messages;
INSERT INTO messages (content) VALUES('Hallo Welt!');
-----------------------------------------------------------------------------------------------------------------------------------------------------------------
DELETE FROM platforms WHERE id = 'de6173ca-05ad-4b28-b872-e61e3285abea';
SELECT * FROM platforms;
-----------------------------------------------------------------------------------------------------------------------------------------------------------------
INSERT INTO platforms (id, env_id, name, api_server, namespace, secret) VALUES('de6173ca-05ad-4b28-b872-e61e3285abea', 'e7ccea48-c007-4ff5-b2fb-74516e77da00', 'demo', 'https://192.168.178.31:8443', 'myproject-test', 'eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJteXByb2plY3QtdGVzdCIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VjcmV0Lm5hbWUiOiJkZXB1dHktdG9rZW4tOHg5ZHEiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiZGVwdXR5Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQudWlkIjoiMjMzMmVlNzEtZDUwNC0xMWViLTk1YjYtMDAxNTVkNjMwMTA4Iiwic3ViIjoic3lzdGVtOnNlcnZpY2VhY2NvdW50Om15cHJvamVjdC10ZXN0OmRlcHV0eSJ9.0KfPfVlTOf0u2EaU6GqPxtjZY9ynzC1d4wGhG30xVd_rGHjWAmoNnuJ8USEPmqsHBo82jugWqNToLmN5envE33SEAe9OgKflDQxoSynWonD6PX4eCW2cASdchXCsAdnvg1GZNd4yXh4P3WZU1zRUZdm0eKi2RbFx1hNF2iKQzIQE5GcQDogZvebYDRTA3SVAxgHNxKX97UOHH6dSA4oeJs5arbh3OUnMwINhpS4v2k-YwQ0nnAyLWbfmdakCUyYK63SEcrR31CHIcwJYRL_syMk0tNdXN1pg6n4w6U7YMczZTeJc6ifIM5ySpqqycvUm-F65pRfSMCcoWEgXLpZLfg');
SELECT * FROM platforms;
-----------------------------------------------------------------------------------------------------------------------------------------------------------------
UPDATE platforms SET name = 'demo1' WHERE id = 'de6173ca-05ad-4b28-b872-e61e3285abea';
SELECT * FROM platforms;
-----------------------------------------------------------------------------------------------------------------------------------------------------------------
