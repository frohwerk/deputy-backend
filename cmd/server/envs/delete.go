package envs

import (
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/pkg/api"
	"github.com/go-chi/chi"
)

func Delete(deleter database.EnvDeleter) http.HandlerFunc {
	delete := func(rw http.ResponseWriter, r *http.Request) (*api.Env, error) {
		id := chi.URLParam(r, "id")
		if id == "" {
			return nil, badRequest("path parameter 'id' may not be empty")
		}
		entity, err := deleter.Delete(id)
		if err != nil {
			return nil, err
		}
		return toApiObject(entity), nil
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		if env, err := delete(rw, r); err != nil {
			writeErrorResponse(rw, err)
		} else {
			writeJsonResponse(rw, env)
		}
	}
}
