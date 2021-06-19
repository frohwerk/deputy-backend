package apps

import (
	"fmt"
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/params"
	"github.com/frohwerk/deputy-backend/internal/request"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
)

func (h *handler) Get(resp http.ResponseWriter, req *http.Request) {
	id := fmt.Sprint(req.Context().Value(params.App))
	envId, _ := request.StringParam(req.URL.Query(), "env")
	before, _ := request.TimeParam(req.URL.Query(), "before")

	fmt.Printf("AppsHandler.Get(%v, %v, %v)\n", id, envId, before)

	var (
		repo   = Repository{h.DB}
		result *App
		err    error
	)

	if before == nil {
		result, err = repo.CurrentView(id, envId)
	} else {
		result, err = repo.History(id, envId, before)
	}

	if err != nil {
		httputil.WriteErrorResponse(resp, err)
	} else {
		httputil.WriteJsonResponse(resp, result)
	}
}
