package components

import "github.com/frohwerk/deputy-backend/internal/database"

type componentHandler struct {
	components  database.ComponentStore
	deployments database.DeploymentStore
}

func NewHandler(
	components database.ComponentStore,
	deployments database.DeploymentStore,
) *componentHandler {
	return &componentHandler{components, deployments}
}
