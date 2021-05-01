package envs

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/pkg/api"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
	"github.com/go-chi/chi"
)

type updateStore interface {
	database.EnvGetter
	database.EnvUpdater
}

func Update(store updateStore) http.HandlerFunc {
	update := func(rw http.ResponseWriter, r *http.Request) (*api.Env, error) {
		id := chi.URLParam(r, "env")
		log.Printf("update env %s", id)
		upd := new(api.EnvAttributes)
		err := json.NewDecoder(r.Body).Decode(upd)
		if err != nil {
			return nil, httputil.BadRequest("could not decode update request for env %s", id)
		}
		entity, err := store.Get(id)
		if err != nil {
			return nil, httputil.NotFound("environment with id '%s' not found", id)
		}
		if upd.Name != "" {
			entity.Name = upd.Name
		}
		entity, err = store.Update(entity)
		if err != nil {
			return nil, err
		}
		return toApiObject(entity), nil
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		if env, err := update(rw, r); err != nil {
			httputil.WriteErrorResponse(rw, err)
		} else {
			httputil.WriteJsonResponse(rw, env)
		}
	}
}
