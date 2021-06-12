CREATE VIEW apps_components_all AS
  SELECT app_id, component_id, updated as valid_from, null as valid_until FROM apps_components
   UNION
  SELECT app_id, component_id, valid_from, valid_until FROM apps_components_history;

CREATE VIEW deployments_all AS
  SELECT component_id, platform_id, updated as valid_from, null as valid_until, image_ref FROM deployments
   UNION
  SELECT component_id, platform_id, valid_from, valid_until, image_ref FROM deployments_history;
