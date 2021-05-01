package envs

import (
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/pkg/api"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
	"github.com/go-chi/chi"
)

type EnvLookup interface {
	Get(id string) (*database.Env, error)
}

func Get(envs EnvLookup) http.HandlerFunc {
	get := func(rw http.ResponseWriter, r *http.Request) (*api.Env, error) {
		id := chi.URLParam(r, "env")
		if id == "" {
			return nil, httputil.BadRequest("path parameter id may not be empty")
		}
		entity, err := envs.Get(id)
		if err != nil {
			return nil, err
		}
		return toApiObject(entity), nil
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		if env, err := get(rw, r); err != nil {
			httputil.WriteErrorResponse(rw, err)
		} else {
			httputil.WriteJsonResponse(rw, env)
		}
	}
}
