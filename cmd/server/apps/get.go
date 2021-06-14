package apps

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/frohwerk/deputy-backend/internal/params"
	"github.com/frohwerk/deputy-backend/internal/request"
	"github.com/frohwerk/deputy-backend/pkg/api"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
)

func (h *handler) Get(resp http.ResponseWriter, req *http.Request) {
	id := fmt.Sprint(req.Context().Value(params.App))
	envId, _ := request.StringParam(req.URL.Query(), "env")
	before, _ := request.FloatParam(req.URL.Query(), "before")

	fmt.Printf("AppsHandler.Get(%v, %v, %v)\n", id, envId, before)

	var (
		result interface{}
		err    error
	)

	if before == 0 {
		result, err = h.currentView(id, envId, resp, req)
	} else {
		result, err = h.history(id, envId, before, resp, req)
	}

	if err != nil {
		httputil.WriteErrorResponse(resp, err)
	} else {
		httputil.WriteJsonResponse(resp, result)
	}
}

type component struct {
	Id       string `json:"id"`
	Name     string `json:"name,omitempty"`
	Image    string `json:"image,omitempty"`
	Deployed string `json:"deployed,omitempty"`
}

type state struct {
	ValidFrom  *time.Time  `json:"validFrom,omitempty"`
	ValidUntil *time.Time  `json:"validUntil,omitempty"`
	Components []component `json:"components"`
}

type app struct {
	Id         string      `json:"id"`
	Name       string      `json:"name,omitempty"`
	Components []component `json:"components,omitempty"`
	History    []state     `json:"history,omitempty"`
}

func (h *handler) history(id, envId string, before float64, resp http.ResponseWriter, req *http.Request) (*app, error) {
	// TODO: Read history using apps_history view
	//  SELECT envs.name AS env_name, apps.name AS app_name, valid_from, components.name AS component_name, image_ref
	//
	// COALESCE(image_ref, ''), COALESCE(ROUND(EXTRACT(EPOCH FROM deployed AT TIME ZONE 'UTC'))::INTEGER, 0)
	//	rows, err := h.DB.Query(`
	//		SELECT apps.id, apps.name, valid_from,
	//		       ROW_NUMBER() OVER (),
	//		       COALESCE(components.id, ''), COALESCE(components.name, ''),
	//			   COALESCE(image_ref, ''), COALESCE(to_char(deployed, 'YYYY-MM-DD HH24:MI:SS.USZ'), '')
	//		  FROM apps_history
	//		  JOIN apps ON apps.id = app_id
	//		  JOIN envs ON envs.id = env_id
	//		  LEFT JOIN components ON components.id = component_id
	//	     WHERE app_id = $1 AND env_id = $2
	//	     ORDER BY 1 DESC, 2
	//		 `, id, envId)
	ts := time.Time{}
	row := h.DB.QueryRow(`SELECT to_timestamp($1) AT TIME ZONE 'UTC'`, before)
	if err := row.Scan(&ts); err != nil {
		fmt.Fprintf(os.Stderr, "failed to convert epoch to timestamp: %v\n", err)
	} else {
		fmt.Printf("epoch -> timestamp: %v -> %v\n", before, ts.String())
	}

	rows, err := h.DB.Query(`
        WITH
        params AS (
          SELECT $1 _app_id, $2 _env_id, to_timestamp($3) AT TIME ZONE 'UTC' _timestamp
        ),
        slice AS (
          SELECT
            (SELECT MAX(valid_from) FROM params, apps_timeline WHERE app_id = _app_id AND env_id = _env_id AND valid_from < _timestamp) AS valid_from,
            (SELECT MIN(valid_from) FROM params, apps_timeline WHERE app_id = _app_id AND env_id = _env_id AND valid_from >= _timestamp) AS valid_until
        )
        SELECT apps.id AS app_id, apps.name AS app_name,
               slice.valid_from, slice.valid_until,
               COALESCE(components.id, '') AS component_id, COALESCE(components.name, '') AS component_name,
               COALESCE(h.image_ref, '') AS image_ref, COALESCE(to_char(h.last_deployment, 'YYYY-MM-DD HH24:MI:SS.USZ'), '') AS last_deployment
        FROM params CROSS JOIN slice
       INNER JOIN apps_history h ON h.app_id = _app_id AND h.env_id = _env_id AND h.valid_from = slice.valid_from
       INNER JOIN apps ON apps.id = h.app_id
        LEFT JOIN components ON components.id = h.component_id
       ORDER BY 3 DESC, 5 ASC
    `, id, envId, before)
	//	 FETCH FIRST 5 ROWS ONLY
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during history query: %v\n", err)
		return nil, err
	}

	var until sql.NullTime
	app := app{History: []state{}}
	//snapshots := []state{}
	prev, curr := time.Time{}, time.Time{}
	for i := -1; rows.Next(); prev = curr {
		fmt.Printf("result row #%v\n", i+2)
		c := component{}

		err := rows.Scan(&app.Id, &app.Name, &curr, &until, &c.Id, &c.Name, &c.Image, &c.Deployed)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error during history row scan: %v\n", err)
			return nil, err
		}

		if !curr.Equal(prev) {
			s := state{ValidFrom: &curr, Components: []component{}}
			if until.Valid {
				s.ValidUntil = &until.Time
			}
			app.History = append(app.History, s)
			i++
		}

		if c.Id != "" {
			app.History[i].Components = append(app.History[i].Components, c)
		}
	}

	return &app, nil
}

func (h *handler) currentView(id, envId string, resp http.ResponseWriter, req *http.Request) (*api.App, error) {
	dbapp, err := h.apps.Get(id)
	if err != nil {
		return nil, err
	}

	dbcomponents, err := h.components.ListAllForApp(id)
	if err != nil {
		return nil, err
	}

	components := make([]api.Component, len(dbcomponents))
	for i, c := range dbcomponents {
		deployments, err := h.listDeployments(c.Id, envId)
		if err != nil {
			return nil, err
		}
		components[i] = api.Component{Id: c.Id, Name: c.Name, Deployments: deployments}
	}

	return &api.App{Id: dbapp.Id, Name: dbapp.Name, Artifacts: components}, nil
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
