package database

import (
	"database/sql"
	"log"

	"github.com/frohwerk/deputy-backend/internal/util"
)

type Component struct {
	Id    string
	Name  string
	Image string
}

type ComponentStore interface {
	ListAll() ([]Component, error)
	ListUnassigned() ([]Component, error)
	ListForApp(id string) ([]Component, error)
}

type componentStore struct {
	db *sql.DB
}

func NewComponentStore(db *sql.DB) *componentStore {
	return &componentStore{db}
}

func (s *componentStore) ListAll() ([]Component, error) {
	rows, err := s.db.Query(`SELECT * FROM components c`)
	if err != nil {
		return nil, err
	}
	defer util.Close(rows, log.Printf)
	return fetchComponents(rows)
}

func (s *componentStore) ListUnassigned() ([]Component, error) {
	rows, err := s.db.Query(`SELECT * FROM components c WHERE NOT EXISTS (SELECT * FROM apps_components ac WHERE ac.component_id = c.id)`)
	if err != nil {
		return nil, err
	}
	defer util.Close(rows, log.Printf)
	return fetchComponents(rows)
}

func (s *componentStore) ListForApp(id string) ([]Component, error) {
	rows, err := s.db.Query(`SELECT c.* FROM apps_components ac JOIN components c ON c.ID = ac.component_id WHERE app_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer util.Close(rows, log.Printf)
	return fetchComponents(rows)
}

func fetchComponents(rows *sql.Rows) ([]Component, error) {
	components := make([]Component, 0)
	for rows.Next() {
		c := Component{}
		if err := rows.Scan(&c.Id, &c.Name, &c.Image); err != nil {
			return nil, err
		}
		components = append(components, c)
	}
	return components, nil
}
