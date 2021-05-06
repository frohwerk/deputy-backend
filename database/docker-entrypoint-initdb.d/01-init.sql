DROP TABLE IF EXISTS envs;
CREATE TABLE envs (
    env_id         VARCHAR(36) PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    env_name       VARCHAR(50) UNIQUE NOT NULL
);

DROP TABLE IF EXISTS platforms;
CREATE TABLE platforms (
    pf_id         VARCHAR(36) PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    pf_env        VARCHAR(36) NOT NULL,
    pf_name       VARCHAR(50),
    pf_api_server VARCHAR(256),
    pf_namespace  VARCHAR(50),
    pf_secret     VARCHAR(2048),
    UNIQUE (pf_env, pf_name),
    FOREIGN KEY (pf_env) REFERENCES envs(env_id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS components;
CREATE TABLE components (
    component_id VARCHAR(36) NOT NULL DEFAULT gen_random_uuid(),
    name         VARCHAR(50) NOT NULL UNIQUE,
    PRIMARY KEY (component_id)
);

CREATE TABLE deployments (
    component_id VARCHAR(36) NOT NULL,
    platform_id  VARCHAR(36) NOT NULL,
    image_ref    VARCHAR(150) NOT NULL,
    updated      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (component_id, platform_id),
    FOREIGN KEY (platform_id) REFERENCES platforms (pf_id) ON DELETE CASCADE,
    FOREIGN KEY (component_id) REFERENCES components (component_id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS apps;
CREATE TABLE apps (
    id VARCHAR(36) PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE
);

DROP TABLE IF EXISTS apps_components;
CREATE TABLE apps_components (
    app_id VARCHAR(36) NOT NULL,
    component_id VARCHAR(36) NOT NULL,
    PRIMARY KEY (app_id, component_id),
    FOREIGN KEY (app_id) REFERENCES apps (id) ON DELETE CASCADE,
    FOREIGN KEY (component_id) REFERENCES components (component_id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS files;
CREATE TABLE files (
    file_id VARCHAR(36) NOT NULL,
    file_digest VARCHAR(71) NOT NULL,
    file_path VARCHAR(4096) NOT NULL,
    file_parent VARCHAR(36) DEFAULT NULL,
    PRIMARY KEY (file_id),
    UNIQUE (file_digest, file_path),
    FOREIGN KEY (file_parent) REFERENCES files(file_id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS images_artifacts;
CREATE TABLE images_artifacts (
    image_id VARCHAR(150) NOT NULL,
    file_id VARCHAR(36) NOT NULL,
    PRIMARY KEY (image_id, file_id),
    FOREIGN KEY (file_id) REFERENCES files(file_id) ON DELETE CASCADE
);

-- TODO: neue tabelle für die component-platform relationship
-- CREATE TABLE platforms_components (
--     platform_id   VARCHAR(36) NOT NULL,
--     component_id   VARCHAR(36) NOT NULL,
--     PRIMARY KEY (platform_id, component_id)
--     FOREIGN KEY (platform_id)  REFERENCES platforms(pf_id) ON DELETE CASCADE,
--     FOREIGN KEY (component_id) REFERENCES components(id)   ON DELETE CASCADE
-- );