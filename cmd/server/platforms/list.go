package platforms

import (
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
	"github.com/go-chi/chi"
)

func List(store database.PlatformLister) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		envId := chi.URLParam(r, "env")
		if platforms, err := store.ListForEnv(envId); err != nil {
			httputil.WriteErrorResponse(rw, err)
		} else {
			for _, p := range platforms {
				p.Secret = ""
			}
			httputil.WriteJsonResponse(rw, platforms)
		}
	}
}
