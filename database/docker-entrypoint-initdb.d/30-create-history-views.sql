CREATE VIEW apps_components_all AS
  SELECT app_id, component_id, updated as valid_from, null as valid_until FROM apps_components
   UNION
  SELECT app_id, component_id, valid_from, valid_until FROM apps_components_history;

CREATE VIEW deployments_all AS
  SELECT component_id, platform_id, updated as valid_from, null as valid_until, image_ref FROM deployments
   UNION
  SELECT component_id, platform_id, valid_from, valid_until, image_ref FROM deployments_history;

DROP VIEW IF EXISTS apps_history;
CREATE OR REPLACE VIEW apps_history AS
  SELECT t.app_id, p.env_id, t.valid_from, c.component_id, d.image_ref
    FROM apps_timeline t
   INNER JOIN platforms p
      ON p.env_id = t.env_id
   INNER JOIN apps_components_all c
      ON c.app_id = t.app_id
     AND c.valid_from <= t.valid_from AND t.valid_from < COALESCE(c.valid_until, CURRENT_TIMESTAMP)
    LEFT JOIN deployments_all d
      ON d.component_id = c.component_id AND d.platform_id = p.id
     AND d.valid_from <= t.valid_from AND t.valid_from < COALESCE(d.valid_until, CURRENT_TIMESTAMP)

--  SELECT t.app_id, t.env_id, t.valid_from, c.component_id, (
--           SELECT d.image_ref FROM platforms p JOIN deployments_all d ON p.id = d.platform_id
--            WHERE p.env_id = t.env_id AND d.component_id = c.component_id
--              AND d.valid_from <= t.valid_from AND t.valid_from < COALESCE(d.valid_until, CURRENT_TIMESTAMP)
--         ) AS image_ref
--    FROM apps_timeline t
--   INNER JOIN apps_components_all c
--      ON c.app_id = t.app_id
--     AND c.valid_from <= t.valid_from AND t.valid_from < COALESCE(c.valid_until, CURRENT_TIMESTAMP);

-- SELECT t.app_id, t.env_id, t.valid_from, c.component_id, d.image_ref
--   FROM apps_timeline t
--  INNER JOIN apps_components_all c
--     ON c.app_id = t.app_id
--    AND c.valid_from <= t.valid_from AND t.valid_from < COALESCE(c.valid_until, CURRENT_TIMESTAMP)
--   LEFT JOIN deployments_all d
--     ON d.component_id = c.component_id
--    AND d.valid_from <= t.valid_from AND t.valid_from < COALESCE(d.valid_until, CURRENT_TIMESTAMP)
--   LEFT JOIN platforms p
--     ON p.env_id = t.env_id
--    AND p.id = d.platform_id
