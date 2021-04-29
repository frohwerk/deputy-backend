package envs

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/pkg/api"
)

var (
	patternName = regexp.MustCompile(`^[A-Za-z](?:[-a-z])*`)
)

type createRequest struct {
	Name string `json:"name,omitempty"`
}

func (r *createRequest) validate() error {
	switch {
	case r.Name == "":
		return badRequest("name attribute may not be empty")
	case !patternName.MatchString(r.Name):
		return badRequest("name attribute may only must begin with a lower-case letter and may only contain lower-case letters and dashes")
	default:
		return nil
	}
}

func Create(store database.EnvCreator) http.HandlerFunc {
	create := func(rw http.ResponseWriter, r *http.Request) (*api.Env, error) {
		cr := new(createRequest)
		err := json.NewDecoder(r.Body).Decode(cr)
		if err != nil {
			return nil, err
		}
		if err := cr.validate(); err != nil {
			return nil, err
		}
		entity, err := store.Create(cr.Name)
		if err != nil {
			return nil, err
		}
		return toApiObject(entity), nil
	}

	return func(rw http.ResponseWriter, r *http.Request) {
		if env, err := create(rw, r); err != nil {
			writeErrorResponse(rw, err)
		} else {
			log.Println("Path:", r.URL.Path)
			uri := fmt.Sprintf("http://%s%s/%s", r.Host, strings.TrimSuffix(r.URL.Path, "/"), env.Id)
			rw.Header().Add("Location", uri)
			rw.WriteHeader(http.StatusCreated)
			writeJsonResponse(rw, env)
		}
	}
}
