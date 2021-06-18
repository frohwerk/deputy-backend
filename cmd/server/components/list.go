package components

import (
	"fmt"
	"net/http"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/epoch"
	"github.com/frohwerk/deputy-backend/internal/request"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
)

func (h *componentHandler) List(rw http.ResponseWriter, r *http.Request) {
	fmt.Println("Incoming request: componentHandler.list")
	components, err := h.list(r.URL.Query())
	if err != nil {
		httputil.WriteErrorResponse(rw, err)
	} else {
		httputil.WriteJsonResponse(rw, components)
	}
}

func (h *componentHandler) list(params map[string][]string) ([]component, error) {
	appId, unassigned := request.StringParam(params, "unassigned")
	envId, _ := request.StringParam(params, "env")

	components, err := h.getComponents(unassigned, appId)
	if err != nil {
		return nil, err
	}

	result := make([]component, len(components))
	for i, c := range components {
		result[i] = component{Id: c.Id, Name: c.Name}

		if envId == "" {
			continue
		}

		deployments, err := h.getDeployments(c.Id, envId)

		switch {
		case err != nil:
			return nil, err
		case len(deployments) == 0:
			continue
		}

		result[i].Image = deployments[0].ImageRef
		result[i].Deployed = epoch.FromTime(&deployments[0].Updated)
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
