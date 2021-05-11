package deployments

import (
	"database/sql"

	"github.com/go-chi/chi"
)

type handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *handler {
	return &handler{db}
}

func (h *handler) Routes(r chi.Router) {
	r.Get("/", h.List)
}
