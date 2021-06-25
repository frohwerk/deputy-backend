CREATE EXTENSION "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

SELECT * FROM pg_available_extensions WHERE name = 'uuid-ossp';
SELECT * FROM pg_extension;

GRANT EXECUTE ON FUNCTION uuid_generate_v4 TO deputy;
SELECT uuid_generate_v4();
SELECT * FROM envs;

DROP TABLE IF EXISTS envs;
CREATE TABLE envs (
    id          VARCHAR(36) NOT NULL DEFAULT uuid_generate_v4(),
    name        VARCHAR(50) NOT NULL UNIQUE,
    order_hint  INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);
