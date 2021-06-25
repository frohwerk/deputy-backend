CREATE EXTENSION "uuid-ossp";
DROP EXTENSION "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

GRANT EXECUTE ON FUNCTION uuid_generate_v4 TO deputy;

SELECT uuid_generate_v4();

SELECT * FROM pg_available_extensions WHERE name = 'uuid-ossp';
SELECT * FROM pg_extension;
