package components

import (
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/pkg/api"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
)

func List(store database.ComponentStore) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		appId, present := stringParam(r.URL.Query(), "unassigned")
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
		components[i] = api.Component{Id: c.Id, Name: c.Name, Image: c.Image, Updated: c.Updated}
	}

	httputil.WriteJsonResponse(resp, components)
}

func booleanParam(queryParams map[string][]string, name string) bool {
	v, ok := queryParams[name]
	return ok && (len(v) == 0 || v[0] == "")
}

func stringParam(queryParams map[string][]string, name string) (string, bool) {
	if v, exists := queryParams[name]; exists && len(v) > 0 {
		return v[0], true
	}
	return "", false
}
