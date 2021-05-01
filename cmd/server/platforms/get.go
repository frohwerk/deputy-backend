package platforms

import (
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
	"github.com/go-chi/chi"
)

func Get(store database.PlatformGetter) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "platform")
		if platform, err := store.Get(id); err != nil {
			httputil.WriteErrorResponse(rw, err)
		} else {
			platform.Secret = ""
			httputil.WriteJsonResponse(rw, platform)
		}
	}
}
