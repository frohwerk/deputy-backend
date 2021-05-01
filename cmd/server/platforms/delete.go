package platforms

import (
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
	"github.com/go-chi/chi"
)

func Delete(store database.PlatformDeleter) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "platform")
		if platform, err := store.Delete(id); err != nil {
			httputil.WriteErrorResponse(rw, err)
		} else {
			httputil.WriteJsonResponse(rw, platform)
		}
	}
}
