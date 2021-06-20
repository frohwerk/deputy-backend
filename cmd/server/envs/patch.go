package envs

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/frohwerk/deputy-backend/pkg/api"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
	"github.com/go-chi/chi"
)

func Patch(db *sql.DB) http.HandlerFunc {
	applyPatch := func(rw http.ResponseWriter, r *http.Request) (*api.Env, error) {
		id := chi.URLParam(r, "env")
		if id == "" {
			return nil, httputil.BadRequest("path parameter id may not be empty")
		}

		patch := api.Env{}
		if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
			return nil, err
		}

		original := api.Env{Id: &id}
		row := db.QueryRow(`SELECT name, order_hint FROM envs WHERE id = $1`, id)
		if err := row.Scan(&original.Name, &original.Order); err != nil {
			return nil, err
		}

		updated := api.Env{Id: original.Id, Name: original.Name, Order: original.Order}
		switch {
		case patch.Name != nil:
			updated.Name = patch.Name
		case patch.Order != nil:
			updated.Order = patch.Order
		}

		if _, err := db.Exec(`UPDATE envs SET name = $1, order_hint = $2 WHERE id = $3`, *updated.Name, *updated.Order, *updated.Id); err != nil {
			return nil, err
		}

		return &updated, nil
	}
	return func(rw http.ResponseWriter, r *http.Request) {
		if env, err := applyPatch(rw, r); err != nil {
			httputil.WriteErrorResponse(rw, err)
		} else {
			httputil.WriteJsonResponse(rw, env)
		}
	}
}
