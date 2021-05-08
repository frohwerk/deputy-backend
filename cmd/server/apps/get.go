package apps

import (
	"fmt"
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/params"
	"github.com/frohwerk/deputy-backend/internal/request"
	"github.com/frohwerk/deputy-backend/pkg/api"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
)

func (h *handler) Get(resp http.ResponseWriter, req *http.Request) {
	id := fmt.Sprint(req.Context().Value(params.App))
	envId, _ := request.StringParam(req.URL.Query(), "env")

	dbapp, err := h.apps.Get(id)
	if err != nil {
		httputil.WriteErrorResponse(resp, err)
		return
	}

	dbcomponents, err := h.components.ListAllForApp(id)
	if err != nil {
		httputil.WriteErrorResponse(resp, err)
		return
	}

	components := make([]api.Component, len(dbcomponents))
	for i, c := range dbcomponents {
		deployments, err := h.listDeployments(c.Id, envId)
		if err != nil {
			httputil.WriteErrorResponse(resp, err)
			return
		}
		components[i] = api.Component{Id: c.Id, Name: c.Name, Deployments: deployments}
	}

	httputil.WriteJsonResponse(resp, api.App{Id: dbapp.Id, Name: dbapp.Name, Artifacts: components})
}

func (h *handler) listDeployments(componentId, envId string) ([]api.Deployment, error) {
	entities, err := h.deployments.ListForEnv(componentId, envId)
	if err != nil {
		return nil, err
	}

	result := make([]api.Deployment, len(entities))
	for i, d := range entities {
		result[i] = api.Deployment{ImageRef: d.ImageRef, Updated: d.Updated}
	}

	return result, nil
}
