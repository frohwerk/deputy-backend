package deployments

import (
	"net/http"
	"time"

	"github.com/frohwerk/deputy-backend/internal/request"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
)

type deployment struct {
	Name    string     `json:"name,omitempty"`
	Image   string     `json:"image,omitempty"`
	Updated *time.Time `json:"updated,omitempty"`
}

func (h *handler) List(rw http.ResponseWriter, r *http.Request) {
	appId, _ := request.StringParam(r.URL.Query(), "app")
	envId, _ := request.StringParam(r.URL.Query(), "env")

	result, err := h.list(appId, envId)

	if err != nil {
		httputil.WriteErrorResponse(rw, err)
	} else {
		httputil.WriteJsonResponse(rw, result)
	}
}

func (h *handler) list(appId, envId string) ([]deployment, error) {
	rows, err := h.db.Query(`
		SELECT d.component_id IS NOT NULL AS DEPLOYED,
			   c.name,
		       COALESCE(d.image_ref, ''), COALESCE(d.updated, TIMESTAMP '0001-01-01 00:00:00+00' AT TIME ZONE 'UTC')
		  FROM apps_components ac
		 CROSS JOIN platforms p
		 INNER JOIN components c ON c.component_id = ac.component_id
		  LEFT JOIN deployments d ON d.component_id = c.component_id AND d.platform_id = p.pf_id
		 WHERE ac.app_id = $1 AND p.pf_env = $2
	`, appId, envId)
	if err != nil {
		return nil, err
	}

	var deployed = false
	result := make([]deployment, 0)
	for i := 0; rows.Next(); i++ {
		result = append(result, deployment{})
		if err := rows.Scan(&deployed, &result[i].Name, &result[i].Image, &result[i].Updated); err != nil {
			return nil, err
		}
		if !deployed {
			result[i].Updated = nil
		}
	}

	return result, nil
}
