package apps

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/frohwerk/deputy-backend/internal/epoch"
)

type Repository struct {
	db *sql.DB
}

type Component struct {
	Id       string       `json:"id"`
	Name     string       `json:"name,omitempty"`
	Image    string       `json:"image,omitempty"`
	Artifact string       `json:"artifact,omitempty"`
	Platform string       `json:"-"`
	Deployed *epoch.Epoch `json:"deployed,omitempty"`
}

type App struct {
	Id         string       `json:"id"`
	Name       string       `json:"name,omitempty"`
	Created    *epoch.Epoch `json:"created,omitempty"`
	ValidFrom  *epoch.Epoch `json:"validFrom,omitempty"`
	ValidUntil *epoch.Epoch `json:"validUntil,omitempty"`
	Components []Component  `json:"components"`
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db}
}

func (h *Repository) History(id, envId string, before *time.Time) (*App, error) {
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
               h.image_ref, platforms.name, h.last_deployment,
			   files.path
        FROM params CROSS JOIN slice
       INNER JOIN apps_history h ON h.app_id = _app_id AND h.env_id = _env_id AND h.valid_from = slice.valid_from
       INNER JOIN apps ON apps.id = h.app_id
	    LEFT JOIN platforms ON platforms.id = h.platform_id
        LEFT JOIN components ON components.id = h.component_id
		LEFT JOIN images_artifacts ia ON ia.image_id = h.image_ref
		LEFT JOIN files on files.id = ia.file_id
	 ORDER BY 3 DESC, 5 ASC
    `, id, envId, before)
}

func (h *Repository) CurrentView(id, envId string) (*App, error) {
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
			   h.image_ref, platforms.name, h.last_deployment,
			   files.path
		  FROM params CROSS JOIN slice
	     INNER JOIN apps_history h ON h.app_id = _app_id AND h.env_id = _env_id AND h.valid_from = slice.valid_from
	     INNER JOIN apps ON apps.id = h.app_id
		  LEFT JOIN platforms ON platforms.id = h.platform_id
		  LEFT JOIN components ON components.id = h.component_id
		  LEFT JOIN images_artifacts ia ON ia.image_id = h.image_ref
		  LEFT JOIN files on files.id = ia.file_id
	     ORDER BY 3 DESC, 5 ASC
	`, id, envId)
}

func (h *Repository) query(query string, args ...interface{}) (*App, error) {
	rows, err := h.db.Query(query, args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during apps query: %v\n", err)
		return nil, err
	}

	app := App{Components: []Component{}}
	//snapshots := []state{}
	for i := 0; rows.Next(); i++ {
		var id, name, image, artifact, platform sql.NullString
		var created, from, until, deployed sql.NullTime
		// fmt.Printf("result row #%v\n", i+1)
		err := rows.Scan(&app.Id, &app.Name, &created, &from, &until, &id, &name, &image, &platform, &deployed, &artifact)
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

		c := Component{}
		if id.Valid {
			c.Id = id.String
			if name.Valid {
				c.Name = name.String
			}
			if image.Valid {
				c.Image = image.String
			}
			if artifact.Valid {
				c.Artifact = artifact.String
			}
			if platform.Valid {
				c.Platform = platform.String
			}
			if deployed.Valid {
				c.Deployed = epoch.FromTime(&deployed.Time)
			}
			app.Components = append(app.Components, c)
		}
	}

	return &app, nil
}
