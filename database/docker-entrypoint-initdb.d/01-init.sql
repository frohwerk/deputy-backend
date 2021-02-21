DROP TABLE IF EXISTS environments;
-- CREATE TABLE environments (
--     id VARCHAR(36) PRIMARY KEY,
--     name VARCHAR(50) UNIQUE
-- );

DROP TABLE IF EXISTS platforms;
-- CREATE TABLE platforms (
--     id VARCHAR(36) PRIMARY KEY,
--     name VARCHAR(50) UNIQUE
-- );

DROP TABLE IF EXISTS apps;
CREATE TABLE apps (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) UNIQUE
);

DROP TABLE IF EXISTS components;
CREATE TABLE components (
    id VARCHAR(36) PRIMARY KEY NOT NULL,
    name VARCHAR(50) UNIQUE NOT NULL,
    updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version VARCHAR(20),
    image VARCHAR(150)
);

DROP TABLE IF EXISTS apps_components;
CREATE TABLE apps_components (
    app_id VARCHAR(36) NOT NULL,
    component_id VARCHAR(36) NOT NULL,
    PRIMARY KEY (app_id, component_id),
    FOREIGN KEY (app_id) REFERENCES apps (id) ON DELETE CASCADE,
    FOREIGN KEY (component_id) REFERENCES components (id) ON DELETE CASCADE
);
