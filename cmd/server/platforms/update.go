package platforms

import (
	"encoding/json"
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/pkg/api"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
	"github.com/go-chi/chi"
)

type updateStore interface {
	database.PlatformGetter
	database.PlatformUpdater
}

type update struct {
	api.Platform
}

func Update(store updateStore) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "platform")

		platform, err := store.Get(id)
		if err != nil {
			httputil.WriteErrorResponse(rw, httputil.NotFound("platform %s does not exist", id))
			return
		}

		changeset := &update{api.Platform{}}
		if err := json.NewDecoder(r.Body).Decode(&changeset.Platform); err != nil {
			httputil.WriteErrorResponse(rw, err)
			return
		}

		changeset.applyTo(platform)

		result, err := store.Update(platform)
		if err != nil {
			httputil.WriteErrorResponse(rw, err)
			return
		}

		httputil.WriteJsonResponse(rw, result)
	}
}

func (u *update) applyTo(platform *api.Platform) {
	if u.Name != "" {
		platform.Name = u.Name
	}
	if u.ServerUri != "" {
		platform.ServerUri = u.ServerUri
	}
	if u.Namespace != "" {
		platform.Namespace = u.Namespace
	}
	if u.Secret != "" {
		platform.Secret = u.Secret
	}
}
