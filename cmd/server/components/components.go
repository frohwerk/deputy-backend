package components

import (
	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/epoch"
)

type component struct {
	Id       string       `json:"id,omitempty"`
	Name     string       `json:"name,omitempty"`
	Image    string       `json:"image,omitempty"`
	Deployed *epoch.Epoch `json:"deployed,omitempty"`
}

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
