package components

import (
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/request"
	"github.com/frohwerk/deputy-backend/pkg/api"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
)

func (h *componentHandler) List(rw http.ResponseWriter, r *http.Request) {
	components, err := h.list(r.URL.Query())
	if err != nil {
		httputil.WriteErrorResponse(rw, err)
	} else {
		httputil.WriteJsonResponse(rw, components)
	}
}

func (h *componentHandler) list(params map[string][]string) ([]api.Component, error) {
	appId, unassigned := request.StringParam(params, "unassigned")
	envId, _ := request.StringParam(params, "env")

	components, err := h.getComponents(unassigned, appId)
	if err != nil {
		return nil, err
	}

	result := make([]api.Component, len(components))
	for i, c := range components {
		result[i] = api.Component{Id: c.Id, Name: c.Name}
		deployments, err := h.getDeployments(c.Id, envId)
		if err != nil {
			return nil, err
		}
		result[i].Deployments = make([]api.Deployment, len(deployments))
		for j, d := range deployments {
			result[i].Deployments[j] = api.Deployment{ImageRef: d.ImageRef, Updated: d.Updated}
		}
	}

	return result, nil
}

func (h *componentHandler) getComponents(unassigned bool, appId string) ([]database.Component, error) {
	switch {
	case !unassigned:
		return h.components.ListAll()
	case appId == "":
		return h.components.ListUnassigned()
	default:
		return h.components.ListUnassignedForApp(appId)
	}
}

func (h *componentHandler) getDeployments(componentId, envId string) ([]database.Deployment, error) {
	return h.deployments.ListForEnv(componentId, envId)
}
