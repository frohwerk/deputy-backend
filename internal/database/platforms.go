package database

import (
	"database/sql"
	"log"

	"github.com/frohwerk/deputy-backend/internal/util"
	"github.com/frohwerk/deputy-backend/pkg/api"
)

// Columns: id, env, name, api_server, namespace, secret

type PlatformLister interface {
	List() ([]api.Platform, error)
	ListForEnv(envId string) ([]api.Platform, error)
}

type PlatformCreator interface {
	Create(envId, name string) (*api.Platform, error)
}

type PlatformGetter interface {
	Get(id string) (*api.Platform, error)
}

type PlatformLookup interface {
	Lookup(envId, name string) (*api.Platform, error)
}

type PlatformUpdater interface {
	Update(p *api.Platform) (*api.Platform, error)
}

type PlatformDeleter interface {
	Delete(id string) (*api.Platform, error)
}

type PlatformStore interface {
	PlatformLister
	PlatformCreator
	PlatformGetter
	PlatformLookup
	PlatformUpdater
	PlatformDeleter
}

type platformStore struct {
	*sql.DB
}

func NewPlatformStore(db *sql.DB) PlatformStore {
	return &platformStore{db}
}

func (s *platformStore) List() ([]api.Platform, error) {
	return s.queryAll(`
		SELECT id, name, COALESCE(api_server, ''), COALESCE(namespace, ''), COALESCE(secret, '')
		FROM platforms
	`)
}

func (s *platformStore) ListForEnv(envId string) ([]api.Platform, error) {
	return s.queryAll(`
		SELECT id, name, COALESCE(api_server, ''), COALESCE(namespace, ''), COALESCE(secret, '')
		FROM platforms
		WHERE env_id = $1
	`, envId)
}

func (s *platformStore) Create(envId, name string) (*api.Platform, error) {
	return s.queryOne(`
		INSERT INTO platforms (env_id, name) VALUES($1, $2)
		RETURNING id, name, COALESCE(api_server, ''), COALESCE(namespace, ''), COALESCE(secret, '')
	`, envId, name)
}

func (s *platformStore) Get(id string) (*api.Platform, error) {
	return s.queryOne(`
		SELECT id, name, COALESCE(api_server, ''), COALESCE(namespace, ''), COALESCE(secret, '')
		FROM platforms
		WHERE id = $1
	`, id)
}

func (s *platformStore) Lookup(envId, name string) (*api.Platform, error) {
	return s.queryOne(`
		SELECT id, name, COALESCE(api_server, ''), COALESCE(namespace, ''), COALESCE(secret, '')
		FROM platforms
		WHERE env_id = $1 AND name = $2
	`, envId, name)
}

func (s *platformStore) Update(p *api.Platform) (*api.Platform, error) {
	return s.queryOne(`
		UPDATE platforms
		SET name = $2, api_server = $3, namespace = $4, secret = $5
		WHERE id = $1
		RETURNING id, name, COALESCE(api_server, ''), COALESCE(namespace, ''), COALESCE(secret, '')
	`, p.Id, p.Name, p.ServerUri, p.Namespace, p.Secret)
}

func (s *platformStore) Delete(id string) (*api.Platform, error) {
	return s.queryOne(`
		DELETE FROM platforms
		WHERE id = $1
		RETURNING id, name, COALESCE(api_server, ''), COALESCE(namespace, ''), COALESCE(secret, '')
	`, id)
}

func (s *platformStore) queryOne(query string, args ...interface{}) (*api.Platform, error) {
	return scanPlatform(s.DB.QueryRow(query, args...))
}

func (s *platformStore) queryAll(query string, args ...interface{}) ([]api.Platform, error) {
	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer util.Close(rows, log.Printf)
	platforms := make([]api.Platform, 0)
	for rows.Next() {
		if platform, err := scanPlatform(rows); err != nil {
			return nil, err
		} else {
			platforms = append(platforms, *platform)
		}
	}
	return platforms, nil
}

func (s *platformStore) exec(query string, args ...interface{}) error {
	_, err := s.DB.Exec(query, args...)
	if err != nil {
		return err
	}
	return nil

}

func scanPlatform(s scanner) (*api.Platform, error) {
	p := api.Platform{}
	if err := s.Scan(&p.Id, &p.Name, &p.ServerUri, &p.Namespace, &p.Secret); err != nil {
		return nil, err
	}
	return &p, nil
}
