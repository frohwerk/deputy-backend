package apps

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/params"
	"github.com/frohwerk/deputy-backend/internal/util"
	"github.com/frohwerk/deputy-backend/pkg/api"
)

func UpdateComponents(as database.AppStore) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		if req.Body == nil {
			resp.WriteHeader(http.StatusBadRequest)
			writeJsonResponse(resp, ErrorResponse{"No entity in request"})
			return
		}
		defer util.Close(req.Body, log.Printf)

		id := fmt.Sprint(req.Context().Value(params.App))
		if id == "" {
			resp.WriteHeader(http.StatusBadRequest)
			writeJsonResponse(resp, ErrorResponse{"Empty id in request uri"})
			return
		}

		components, err := decodeComponents(resp, req.Body)
		if err != nil {
			writeErrorResponse(resp, err)
			return
		}

		componentIds := make([]string, len(components))
		for i, v := range components {
			componentIds[i] = v.Id
		}

		if err := as.UpdateComponents(req.Context(), id, componentIds); err != nil {
			writeErrorResponse(resp, err)
			return
		}
	}
}

func decodeComponents(resp http.ResponseWriter, r io.ReadCloser) ([]api.Component, error) {
	components := make([]api.Component, 0)
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&components); err != nil {
		return nil, err
	} else {
		return components, nil
	}
}
