package database

import (
	"database/sql"
	"log"
	"time"

	"github.com/frohwerk/deputy-backend/internal/database/sqlstate"
	"github.com/frohwerk/deputy-backend/internal/util"
)

type Component struct {
	Id      string
	Name    string
	Updated time.Time
	Version string
	Image   string
}

type ComponentStore interface {
	Create(string) (*Component, error)
	CreateIfAbsent(string) (*Component, error)
	SetImage(string, string) (*Component, error)
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

func (s *componentStore) Create(name string) (*Component, error) {
	return s.selectComponent(`
		INSERT INTO components (name)
		VALUES ($1)
		RETURNING id, name, updated, COALESCE(version, ''), COALESCE(image, '')
	`, name)
}

func (s *componentStore) CreateIfAbsent(name string) (*Component, error) {
	c, err := s.Create(name)
	switch {
	case sqlstate.UniqueViolation(err):
		return s.GetByName(name)
	case err != nil:
		return nil, err
	default:
		return c, nil
	}
}

func (s *componentStore) SetImage(name string, image string) (*Component, error) {
	return s.selectComponent(`
		UPDATE components
		SET image = $2
		WHERE name = $1
		RETURNING id, name, updated, COALESCE(version, ''), COALESCE(image, '')
	`, name, image)
}

func (s *componentStore) GetByName(name string) (*Component, error) {
	return s.selectComponent(`
		SELECT id, name, updated, COALESCE(version, ''), COALESCE(image, '')
		FROM components
		WHERE name = $1
	`, name)
}

func (s *componentStore) ListAll() ([]Component, error) {
	return s.selectComponents(`
		SELECT id, name, updated, COALESCE(version, ''), COALESCE(image, '')
		FROM components
	`)
}

func (s *componentStore) ListUnassigned() ([]Component, error) {
	return s.selectComponents(`
		SELECT id, name, updated, COALESCE(version, ''), COALESCE(image, '')
		FROM components c
	  	WHERE NOT EXISTS (SELECT * FROM apps_components ac WHERE ac.component_id = c.id)
	`)
}

func (s *componentStore) ListForApp(id string) ([]Component, error) {
	return s.selectComponents(`
		SELECT c.id, c.name, c.updated, COALESCE(c.version, ''), COALESCE(c.image, '')
		FROM components c
	  	WHERE NOT EXISTS (SELECT * FROM apps_components ac WHERE ac.component_id = c.id)
	`)
}

func (s *componentStore) selectComponent(query string, args ...interface{}) (*Component, error) {
	row := s.db.QueryRow(query, args...)
	c := &Component{}
	if err := row.Scan(&c.Id, &c.Name, &c.Updated, &c.Version, &c.Image); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *componentStore) selectComponents(query string, args ...interface{}) ([]Component, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer util.Close(rows, log.Printf)
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

func (c Component) modified(other Component) bool {
	return true
}
