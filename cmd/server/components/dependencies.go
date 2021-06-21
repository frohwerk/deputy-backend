package components

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/frohwerk/deputy-backend/internal/util"
	"github.com/frohwerk/deputy-backend/pkg/httputil"
	"github.com/go-chi/chi"
)

type marker struct {
	Id      string `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

type store interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type dependencies struct {
	Additions []string `json:"additions,omitempty"`
	Removals  []string `json:"removals,omitempty"`
}

func (deps *dependencies) String() string {
	buf, err := json.Marshal(deps)
	if err != nil {
		return fmt.Sprint(err)
	}
	return string(buf)
}

func (h *componentHandler) GetDependencies(rw http.ResponseWriter, r *http.Request) {
	if result, err := h.getDependencies(r.Context(), chi.URLParam(r, "component")); err != nil {
		httputil.WriteErrorResponse(rw, err)
	} else {
		httputil.WriteJsonResponse(rw, result)
	}
}

func (h *componentHandler) PatchDependencies(rw http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "component")
	deps := &dependencies{}

	if err := json.NewDecoder(r.Body).Decode(deps); err != nil {
		httputil.WriteErrorResponse(rw, err)
		return
	}

	fmt.Printf("PATCH /api/components/%s/dependencies\n", id)
	fmt.Printf("%s\n", deps)
	for _, dependent := range deps.Additions {
		deps, err := h.findAll(dependent)
		if err != nil {
			httputil.WriteErrorResponse(rw, err)
			return
		}
		if deps.Contains(id) {
			rw.WriteHeader(http.StatusBadRequest)
			httputil.WriteJsonResponse(rw, &marker{Id: dependent, Message: "circular dependency"})
			return
		}
	}

	result, err := h.updateDependencies(r.Context(), id, deps)
	if err != nil {
		httputil.WriteErrorResponse(rw, err)
		return
	}

	for _, c := range result {
		fmt.Printf("dependency: %s, %s\n", c.Id, c.Name)
	}

	httputil.WriteJsonResponse(rw, result)
}

func (h *componentHandler) getDependencies(ctx context.Context, id string) ([]component, error) {
	result := []component{}

	rows, err := h.db.QueryContext(ctx, `SELECT c.id, c.name FROM dependencies d JOIN components c ON c.id = d.depends_on WHERE d.id = $1`, id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		c := component{}
		err := rows.Scan(&c.Id, &c.Name)
		if err != nil {
			return nil, err
		}
		result = append(result, c)
	}

	return result, nil
}

func (h *componentHandler) findAll(id string) (util.Set, error) {
	result := make(util.Set)

	pending := []string{id}

	fmt.Println("Searching dependencies for", id)
	for len(pending) > 0 {
		id, pending = pending[0], pending[1:]
		rows, err := h.db.Query(`SELECT depends_on FROM dependencies WHERE id = $1`, id)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var dependent string
			err := rows.Scan(&dependent)
			if err != nil {
				return nil, err
			}

			fmt.Println("Found dependency:", dependent)

			if !result.Contains(dependent) {
				result.Put(dependent)
				fmt.Println("Queueing dependency lookup for", dependent)
				pending = append(pending, dependent)
			}
		}
	}

	return result, nil
}

func (h *componentHandler) updateDependencies(ctx context.Context, id string, deps *dependencies) ([]component, error) {
	tx, err := h.db.BeginTx(ctx, &sql.TxOptions{})

	if err != nil {
		return rollback(tx, err)
	}

	for _, addition := range deps.Additions {
		fmt.Printf("INSERT INTO dependencies (id, depends_on) VALUES ('%s', '%s') ON CONFLICT DO NOTHING\n", id, addition)
		_, err := tx.ExecContext(ctx, `INSERT INTO dependencies (id, depends_on) VALUES ($1, $2) ON CONFLICT DO NOTHING`, id, addition)
		if err != nil {
			return rollback(tx, err)
		}
	}

	for _, removal := range deps.Removals {
		fmt.Printf("DELETE FROM dependencies WHERE id = '%s' AND depends_on = '%s'\n", id, removal)
		_, err := tx.ExecContext(ctx, `DELETE FROM dependencies WHERE id = $1 AND depends_on = $2`, id, removal)
		if err != nil {
			return rollback(tx, err)
		}
	}

	tx.Commit()

	return h.getDependencies(ctx, id)
}

func rollback(tx *sql.Tx, cause error) ([]component, error) {
	err := tx.Rollback()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error during transaction rollback: %s\n", err)
	}
	return nil, cause
}
