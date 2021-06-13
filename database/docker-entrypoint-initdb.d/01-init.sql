DROP TABLE IF EXISTS envs;
CREATE TABLE envs (
    id         VARCHAR(36) NOT NULL DEFAULT gen_random_uuid(),
    name       VARCHAR(50) NOT NULL UNIQUE,
    PRIMARY KEY (id)
);

DROP TABLE IF EXISTS platforms;
CREATE TABLE platforms (
    id         VARCHAR(36) PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    env_id     VARCHAR(36) NOT NULL,
    name       VARCHAR(50),
    api_server VARCHAR(256),
    namespace  VARCHAR(50),
    secret     VARCHAR(2048),
    UNIQUE (env_id, name),
    FOREIGN KEY (env_id) REFERENCES envs (id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS components;
CREATE TABLE components (
    id    VARCHAR(36) NOT NULL DEFAULT gen_random_uuid(),
    name  VARCHAR(50) NOT NULL UNIQUE,
    PRIMARY KEY (id)
);

-- Deployments entity
CREATE TABLE deployments (
    component_id VARCHAR(36) NOT NULL,
    platform_id  VARCHAR(36) NOT NULL,
    updated      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    image_ref    VARCHAR(150) NOT NULL,
    PRIMARY KEY (component_id, platform_id),
    FOREIGN KEY (platform_id) REFERENCES platforms (id) ON DELETE CASCADE,
    FOREIGN KEY (component_id) REFERENCES components (id) ON DELETE CASCADE
);

CREATE TABLE deployments_history (
    component_id VARCHAR(36) NOT NULL,
    platform_id  VARCHAR(36) NOT NULL,
    valid_from   TIMESTAMP NOT NULL,
    valid_until  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    image_ref    VARCHAR(150) NOT NULL,
    PRIMARY KEY (component_id, platform_id, valid_from),
    FOREIGN KEY (platform_id) REFERENCES platforms (id) ON DELETE CASCADE,
    FOREIGN KEY (component_id) REFERENCES components (id) ON DELETE CASCADE
);

-- Apps entity
DROP TABLE IF EXISTS apps;
CREATE TABLE apps (
    id    VARCHAR(36) NOT NULL DEFAULT gen_random_uuid(),
    name  VARCHAR(50) UNIQUE,
    PRIMARY KEY (id)
);

DROP TABLE IF EXISTS apps_timeline;
CREATE TABLE apps_timeline (
  app_id       VARCHAR(36) NOT NULL,
  env_id       VARCHAR(36) NOT NULL,
  valid_from   TIMESTAMP NOT NULL,
  FOREIGN KEY (app_id) REFERENCES apps (id) ON DELETE CASCADE,
  FOREIGN KEY (env_id) REFERENCES envs (id) ON DELETE CASCADE,
  PRIMARY KEY (app_id, env_id, valid_from)
);
-- TODO:
-- Write apps_timeline entry on updates for apps_components or deployments

-- Relationship apps<->components
DROP TABLE IF EXISTS apps_components;
CREATE TABLE apps_components (
    app_id        VARCHAR(36) NOT NULL,
    component_id  VARCHAR(36) NOT NULL,
    updated       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (app_id, component_id),
    FOREIGN KEY (app_id) REFERENCES apps (id) ON DELETE CASCADE,
    FOREIGN KEY (component_id) REFERENCES components (id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS apps_components_history;
CREATE TABLE apps_components_history (
    app_id        VARCHAR(36) NOT NULL,
    component_id  VARCHAR(36) NOT NULL,
    valid_from    TIMESTAMP NOT NULL,
    valid_until   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (app_id) REFERENCES apps (id) ON DELETE CASCADE,
    FOREIGN KEY (component_id) REFERENCES components (id) ON DELETE CASCADE,
    PRIMARY KEY (app_id, component_id, valid_from)
);

-- Entity files
DROP TABLE IF EXISTS files;
CREATE TABLE files (
    id         VARCHAR(36) NOT NULL,
    digest     VARCHAR(71) NOT NULL,
    path       VARCHAR(4096) NOT NULL,
    parent_id  VARCHAR(36) DEFAULT NULL,
    FOREIGN KEY (parent_id) REFERENCES files (id) ON DELETE CASCADE,
    UNIQUE (digest, path),
    PRIMARY KEY (id)
);

DROP TABLE IF EXISTS images_artifacts;
CREATE TABLE images_artifacts (
    image_id  VARCHAR(150) NOT NULL,
    file_id   VARCHAR(36) NOT NULL,
    PRIMARY KEY (image_id, file_id),
    FOREIGN KEY (file_id) REFERENCES files (id) ON DELETE CASCADE
);
