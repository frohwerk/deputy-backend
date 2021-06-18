package components

import (
	"database/sql"

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
	db          *sql.DB
	components  database.ComponentStore
	deployments database.DeploymentStore
}

func NewHandler(
	db *sql.DB,
	components database.ComponentStore,
	deployments database.DeploymentStore,
) *componentHandler {
	return &componentHandler{db, components, deployments}
}
