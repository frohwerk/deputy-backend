package apps

import (
	"fmt"
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/params"
	"github.com/frohwerk/deputy-backend/pkg/api"
)

func Get(as database.AppStore, cs database.ComponentStore) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		id := fmt.Sprint(req.Context().Value(params.App))

		dbapp, err := as.Get(id)
		if err != nil {
			writeErrorResponse(resp, err)
			return
		}

		dbcomponents, err := cs.ListForApp(id)
		if err != nil {
			writeErrorResponse(resp, err)
			return
		}

		components := make([]api.Component, len(dbcomponents))
		for i, c := range dbcomponents {
			components[i] = api.Component{Id: c.Id, Name: c.Name, Image: c.Image}
		}

		writeJsonResponse(resp, api.App{Id: dbapp.Id, Name: dbapp.Name, Artifacts: components})
	}
}
