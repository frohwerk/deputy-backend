package apps

import (
	"io"
	"net/http"

	"github.com/frohwerk/deputy-backend/pkg/api"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
	"github.com/go-chi/chi"
)

func (h *handler) Patch(rw http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "app")
	app, err := h.patch(id, r.Body)
	if err != nil {
		httputil.WriteErrorResponse(rw, err)
	} else {
		httputil.WriteJsonResponse(rw, app)
	}
}

func (h *handler) patch(id string, body io.ReadCloser) (*api.App, error) {
	attrs, err := tryDecode(body)
	if err != nil {
		return nil, err
	}

	app, err := h.apps.Get(id)
	if err != nil {
		return nil, err
	}

	if len(attrs.Name) > 0 {
		app.Name = attrs.Name
	}

	result, err := h.apps.Update(app)
	if err != nil {
		return nil, err
	}

	return &api.App{Id: result.Id, Name: result.Name}, nil
}
