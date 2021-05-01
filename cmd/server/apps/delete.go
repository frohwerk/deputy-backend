package apps

import (
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
	"github.com/go-chi/chi"
)

type appsDeleter interface {
	database.AppDeleter
}

func Delete(store appsDeleter) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "app")
		if app, err := store.Delete(id); err != nil {
			httputil.WriteErrorResponse(rw, err)
		} else {
			httputil.WriteJsonResponse(rw, app)
		}
	}
}
