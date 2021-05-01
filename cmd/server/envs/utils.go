package envs

import (
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/pkg/api"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
)

func toApiObject(env *database.Env) *api.Env {
	return &api.Env{Id: env.Id, Name: env.Name}
}

func sendResult(entity *database.Env, err error, rw http.ResponseWriter) {
	if err != nil {
		httputil.WriteErrorResponse(rw, err)
		return
	}

	httputil.WriteJsonResponse(rw, entity)
}

func sendResults(entities []database.Env, err error, rw http.ResponseWriter) {
	if err != nil {
		httputil.WriteErrorResponse(rw, err)
		return
	}

	envs := make([]api.Env, len(entities))
	for i, v := range entities {
		envs[i] = api.Env{Id: v.Id, Name: v.Name}
	}

	httputil.WriteJsonResponse(rw, envs)
}
