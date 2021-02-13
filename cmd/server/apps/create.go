package apps

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/util"
	"github.com/frohwerk/deputy-backend/pkg/api"
)

func Create(as database.AppStore) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		if req.Body == nil {
			resp.WriteHeader(http.StatusBadRequest)
			writeJsonResponse(resp, ErrorResponse{"Missing api.Application entity in request"})
			return
		}
		defer util.Close(req.Body, log.Printf)

		app, err := tryDecode(resp, req.Body)
		if err != nil {
			writeErrorResponse(resp, err)
			return
		}

		dbapp, err := as.Create(database.App{Name: app.Name})
		if err != nil {
			writeErrorResponse(resp, err)
			return
		}

		// TODO: Fix format errors in url
		resp.Header().Add("Location", fmt.Sprintf("%s://%s:%s%s/%s", req.URL.Scheme, req.URL.Host, req.URL.Port(), req.URL.Path, url.PathEscape(dbapp.Id)))
		resp.WriteHeader(http.StatusCreated)
		writeJsonResponse(resp, api.App{Name: dbapp.Name})
	}
}

func tryDecode(resp http.ResponseWriter, r io.ReadCloser) (*api.App, error) {
	app := new(api.App)
	dec := json.NewDecoder(r)
	if err := dec.Decode(app); err != nil {
		return nil, err
	} else {
		return app, nil
	}
}
