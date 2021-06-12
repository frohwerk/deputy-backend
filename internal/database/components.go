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
	ListAllForApp(id string) ([]Component, error)
	ListUnassigned() ([]Component, error)
	ListUnassignedForApp(id string) ([]Component, error)
}

type componentStore struct {
	db *sql.DB
}

func NewComponentStore(db *sql.DB) ComponentStore {
	return &componentStore{db}
}

func (s *componentStore) Create(name string) (*Component, error) {
	return s.selectComponent(`
		INSERT INTO components (name)
		VALUES ($1)
		RETURNING id, name
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
		RETURNING id, name
	`, name, image)
}

func (s *componentStore) GetByName(name string) (*Component, error) {
	return s.selectComponent(`
		SELECT id, name
		FROM components
		WHERE name = $1
	`, name)
}

func (s *componentStore) ListAll() ([]Component, error) {
	return s.selectComponents(`
		SELECT id, name
		FROM components
	`)
}

func (s *componentStore) ListUnassigned() ([]Component, error) {
	return s.selectComponents(`
		SELECT id, name
		FROM components c
	  	WHERE NOT EXISTS (SELECT * FROM apps_components ac WHERE ac.component_id = c.id)
	`)
}

func (s *componentStore) ListAllForApp(id string) ([]Component, error) {
	return s.selectComponents(`
		SELECT c.id, c.name
 	    FROM apps_components r
 		JOIN components c
 		ON c.id = r.component_id
		WHERE r.app_id = $1
	`, id)
}

func (s *componentStore) ListUnassignedForApp(id string) ([]Component, error) {
	return s.selectComponents(`
		SELECT c.id, c.name
		FROM components c
		WHERE NOT EXISTS (SELECT * FROM apps_components r WHERE r.component_id = c.id and r.app_id = $1)
	`, id)
}

func (s *componentStore) selectComponent(query string, args ...interface{}) (*Component, error) {
	row := s.db.QueryRow(query, args...)
	if c, err := scanComponent(row); err != nil {
		return nil, err
	} else {
		return c, nil
	}
}

func (s *componentStore) selectComponents(query string, args ...interface{}) ([]Component, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer util.Close(rows, log.Printf)
	components := make([]Component, 0)
	for rows.Next() {
		if c, err := scanComponent(rows); err != nil {
			return nil, err
		} else {
			components = append(components, *c)
		}
	}
	return components, nil
}

func scanComponent(s scanner) (*Component, error) {
	c := Component{}
	if err := s.Scan(&c.Id, &c.Name); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c Component) modified(other Component) bool {
	return true
}
