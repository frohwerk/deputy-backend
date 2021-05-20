package tasks

import (
	"database/sql"

	"github.com/go-chi/chi"
)

type handler struct {
	db *sql.DB
}

func CreateHandler(db *sql.DB) *handler {
	return &handler{db}
}

func (h *handler) Routes(r chi.Router) {
	r.Post("/copy", h.doCopy)
}
