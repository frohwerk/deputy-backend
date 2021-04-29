package envs

import (
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/database"
)

func List(store database.EnvLister) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		entities, err := store.List()
		sendResults(entities, err, rw)
	}
}
