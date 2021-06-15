package apps

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/frohwerk/deputy-backend/internal/epoch"
	"github.com/frohwerk/deputy-backend/internal/params"
	"github.com/frohwerk/deputy-backend/internal/request"
	"github.com/frohwerk/deputy-backend/pkg/api"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
)

func (h *handler) Get(resp http.ResponseWriter, req *http.Request) {
	id := fmt.Sprint(req.Context().Value(params.App))
	envId, _ := request.StringParam(req.URL.Query(), "env")
	before, _ := request.TimeParam(req.URL.Query(), "before")

	fmt.Printf("AppsHandler.Get(%v, %v, %v)\n", id, envId, before)

	var (
		result interface{}
		err    error
	)

	if before == nil {
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
	Id       string       `json:"id"`
	Name     string       `json:"name,omitempty"`
	Image    string       `json:"image,omitempty"`
	Deployed *epoch.Epoch `json:"deployed,omitempty"`
}

type state struct {
	ValidFrom  *epoch.Epoch `json:"validFrom,omitempty"`
	ValidUntil *epoch.Epoch `json:"validUntil,omitempty"`
	Components []component  `json:"components"`
}

type app struct {
	Id         string       `json:"id"`
	Name       string       `json:"name,omitempty"`
	Created    *epoch.Epoch `json:"created,omitempty"`
	ValidFrom  *epoch.Epoch `json:"validFrom,omitempty"`
	ValidUntil *epoch.Epoch `json:"validUntil,omitempty"`
	Components []component  `json:"components"`
}

func (h *handler) history(id, envId string, before *time.Time, resp http.ResponseWriter, req *http.Request) (*app, error) {
	return h.query(`
        WITH
        params AS (
          SELECT $1 _app_id, $2 _env_id, $3::TIMESTAMP _timestamp
        ),
        slice AS (
          SELECT
		    (SELECT MIN(valid_from) FROM params, apps_timeline WHERE app_id = _app_id AND env_id = _env_id) AS created,
            (SELECT MAX(valid_from) FROM params, apps_timeline WHERE app_id = _app_id AND env_id = _env_id AND valid_from < _timestamp) AS valid_from,
            (SELECT MIN(valid_from) FROM params, apps_timeline WHERE app_id = _app_id AND env_id = _env_id AND valid_from >= _timestamp) AS valid_until
        )
        SELECT apps.id, apps.name,
               slice.created, slice.valid_from, slice.valid_until,
               components.id, components.name,
               h.image_ref, h.last_deployment
        FROM params CROSS JOIN slice
       INNER JOIN apps_history h ON h.app_id = _app_id AND h.env_id = _env_id AND h.valid_from = slice.valid_from
       INNER JOIN apps ON apps.id = h.app_id
        LEFT JOIN components ON components.id = h.component_id
       ORDER BY 3 DESC, 5 ASC
    `, id, envId, before)
}

func (h *handler) currentView(id, envId string, resp http.ResponseWriter, req *http.Request) (*app, error) {
	return h.query(`
		WITH
		params AS (
		  SELECT $1 _app_id, $2 _env_id
		),
		slice AS (
		   SELECT MIN(valid_from) AS created, MAX(valid_from) AS valid_from, NULL::TIMESTAMP AS valid_until
		     FROM params, apps_timeline
			WHERE app_id = _app_id AND env_id = _env_id
		)
		SELECT apps.id, apps.name,
			   slice.created, slice.valid_from, slice.valid_until,
			   components.id, components.name,
			   h.image_ref, h.last_deployment
		  FROM params CROSS JOIN slice
	     INNER JOIN apps_history h ON h.app_id = _app_id AND h.env_id = _env_id AND h.valid_from = slice.valid_from
	     INNER JOIN apps ON apps.id = h.app_id
		  LEFT JOIN components ON components.id = h.component_id
	     ORDER BY 3 DESC, 5 ASC
	`, id, envId)
}

func (h *handler) query(query string, args ...interface{}) (*app, error) {
	rows, err := h.DB.Query(query, args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during apps query: %v\n", err)
		return nil, err
	}

	app := app{Components: []component{}}
	//snapshots := []state{}
	for i := 0; rows.Next(); i++ {
		var id, name, image sql.NullString
		var created, from, until, deployed sql.NullTime
		fmt.Printf("result row #%v\n", i+1)
		err := rows.Scan(&app.Id, &app.Name, &created, &from, &until, &id, &name, &image, &deployed)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error during history row scan: %v\n", err)
			return nil, err
		}

		if i == 0 {
			if created.Valid {
				app.Created = epoch.FromTime(&created.Time)
			}
			if from.Valid {
				app.ValidFrom = epoch.FromTime(&from.Time)
			}
			if until.Valid {
				app.ValidUntil = epoch.FromTime(&until.Time)
			}
		}

		c := component{}
		if id.Valid {
			c.Id = id.String
			if name.Valid {
				c.Name = name.String
			}
			if image.Valid {
				c.Image = image.String
			}
			if deployed.Valid {
				c.Deployed = epoch.FromTime(&deployed.Time)
			}
			app.Components = append(app.Components, c)
		}
	}

	return &app, nil
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
