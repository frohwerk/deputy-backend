package components

import (
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/pkg/api"
)

func List(store database.ComponentStore) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		if truthy(req.URL.Query(), "unassigned") {
			entities, err := store.ListUnassigned()
			sendResult(entities, err, resp)
		} else {
			entities, err := store.ListAll()
			sendResult(entities, err, resp)
		}
	}
}

func sendResult(entities []database.Component, err error, resp http.ResponseWriter) {
	if err != nil {
		writeErrorResponse(resp, err)
		return
	}

	components := make([]api.Component, len(entities))
	for i, c := range entities {
		components[i] = api.Component{Id: c.Id, Name: c.Name, Image: c.Image}
	}

	writeJsonResponse(resp, components)
}

func truthy(queryParams map[string][]string, name string) bool {
	v, ok := queryParams[name]
	return ok || len(v) > 0 && v[0] == "false"
}
