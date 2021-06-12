DROP TRIGGER write_deployments_history ON deployments;
DROP FUNCTION write_deployments_history;
DROP VIEW apps_history;
DROP TABLE apps_timeline;
CREATE TABLE apps_timeline (
  app_id       VARCHAR(36) NOT NULL,
  env_id       VARCHAR(36) NOT NULL,
  valid_from   TIMESTAMP NOT NULL,
  FOREIGN KEY (app_id) REFERENCES apps (id) ON DELETE CASCADE,
  FOREIGN KEY (env_id) REFERENCES envs (id) ON DELETE CASCADE,
  PRIMARY KEY (app_id, env_id, valid_from)
);

CREATE VIEW apps_history AS
SELECT t.app_id, t.env_id, /*t.iteration,*/ t.valid_from, /*t.valid_until,*/ c.component_id, d.image_ref --, d.valid_from, d.valid_until
  FROM apps_timeline t
 INNER JOIN apps_components_all c
    ON c.app_id = t.app_id
   AND c.valid_from <= t.valid_from AND t.valid_from < COALESCE(c.valid_until, CURRENT_TIMESTAMP)
  LEFT JOIN deployments_all d
    ON d.component_id = c.component_id
   AND d.valid_from <= t.valid_from AND t.valid_from < COALESCE(d.valid_until, CURRENT_TIMESTAMP)
-- WHERE t.app_id = 'demo'
-- ORDER BY t.valid_from, c.component_id, d.valid_from
;
