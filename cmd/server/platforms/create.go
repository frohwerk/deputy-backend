package platforms

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/pkg/api"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
	"github.com/go-chi/chi"
)

type platformCreator interface {
	Create(name string) (*api.Platform, error)
}

type createRequest struct {
	Name string `json:"name,omitempty"`
}

func (r *createRequest) validate() error {
	return nil
}

func Create(store database.PlatformCreator) http.HandlerFunc {
	create := func(rw http.ResponseWriter, r *http.Request) (*api.Platform, error) {
		envId := chi.URLParam(r, "env")
		cr := new(createRequest)
		err := json.NewDecoder(r.Body).Decode(cr)
		if err != nil {
			return nil, err
		}
		if err := cr.validate(); err != nil {
			return nil, err
		}
		platform, err := store.Create(envId, cr.Name)
		if err != nil {
			return nil, err
		}
		return platform, nil
	}

	return func(rw http.ResponseWriter, r *http.Request) {
		if env, err := create(rw, r); err != nil {
			httputil.WriteErrorResponse(rw, err)
		} else {
			log.Println("Path:", r.URL.Path)
			uri := fmt.Sprintf("http://%s%s/%s", r.Host, strings.TrimSuffix(r.URL.Path, "/"), env.Id)
			rw.Header().Add("Location", uri)
			rw.WriteHeader(http.StatusCreated)
			httputil.WriteJsonResponse(rw, env)
		}
	}
}
