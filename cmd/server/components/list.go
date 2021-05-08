package components

import (
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/request"
	"github.com/frohwerk/deputy-backend/pkg/api"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
)

func List(store database.ComponentStore) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		appId, present := request.StringParam(r.URL.Query(), "unassigned")
		switch {
		case !present:
			entities, err := store.ListAll()
			sendResult(entities, err, rw)
		case appId == "":
			entities, err := store.ListUnassigned()
			sendResult(entities, err, rw)
		default:
			entities, err := store.ListUnassignedForApp(appId)
			sendResult(entities, err, rw)
		}
	}
}

func sendResult(entities []database.Component, err error, resp http.ResponseWriter) {
	if err != nil {
		httputil.WriteErrorResponse(resp, err)
		return
	}

	components := make([]api.Component, len(entities))
	for i, c := range entities {
		components[i] = api.Component{Id: c.Id, Name: c.Name}
	}

	httputil.WriteJsonResponse(resp, components)
}
