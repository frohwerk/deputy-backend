package apps

import (
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/pkg/api"
)

func List(as database.AppStore) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		dbapps, err := as.ListAll()
		if err != nil {
			writeErrorResponse(resp, err)
			return
		}

		apps := make([]api.App, len(dbapps))
		for i, a := range dbapps {
			apps[i] = api.App{Id: a.Id, Name: a.Name}
		}

		writeJsonResponse(resp, apps)
	}
}
