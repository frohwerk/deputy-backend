package apps

import (
	"database/sql"

	"github.com/frohwerk/deputy-backend/internal/database"
)

type handler struct {
	*sql.DB
	apps        database.AppStore
	components  database.ComponentStore
	deployments database.DeploymentStore
	envs        database.EnvStore
	platforms   database.PlatformStore
}

func NewHandler(
	db *sql.DB,
	apps database.AppStore,
	components database.ComponentStore,
	deployments database.DeploymentStore,
	envs database.EnvStore,
	platforms database.PlatformStore,
) *handler {
	return &handler{db, apps, components, deployments, envs, platforms}
}
