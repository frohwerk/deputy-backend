package database

import (
	"database/sql"
	"log"

	"github.com/frohwerk/deputy-backend/internal/util"
	"github.com/frohwerk/deputy-backend/pkg/api"
)

// Columns: pf_id, pf_env, pf_name, pf_api_server, pf_namespace, pf_secret

type PlatformLister interface {
	ListForEnv(envId string) ([]api.Platform, error)
}

type PlatformCreator interface {
	Create(envId, name string) (*api.Platform, error)
}

type PlatformGetter interface {
	Get(id string) (*api.Platform, error)
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
	PlatformUpdater
	PlatformDeleter
}

type platformStore struct {
	*sql.DB
}

func NewPlatformStore(db *sql.DB) PlatformStore {
	return &platformStore{db}
}

func (s *platformStore) ListForEnv(envId string) ([]api.Platform, error) {
	return s.queryAll(`
		SELECT pf_id, pf_name, COALESCE(pf_api_server, ''), COALESCE(pf_namespace, ''), COALESCE(pf_secret, '')
		FROM platforms
		WHERE pf_env = $1
	`, envId)
}

func (s *platformStore) Create(envId, name string) (*api.Platform, error) {
	return s.queryOne(`
		INSERT INTO platforms (pf_env, pf_name) VALUES($1, $2)
		RETURNING pf_id, pf_name, COALESCE(pf_api_server, ''), COALESCE(pf_namespace, ''), COALESCE(pf_secret, '')
	`, envId, name)
}

func (s *platformStore) Get(id string) (*api.Platform, error) {
	return s.queryOne(`
		SELECT pf_id, pf_name, COALESCE(pf_api_server, ''), COALESCE(pf_namespace, ''), COALESCE(pf_secret, '')
		FROM platforms
		WHERE pf_id = $1
	`, id)
}

func (s *platformStore) Update(p *api.Platform) (*api.Platform, error) {
	return s.queryOne(`
		UPDATE platforms
		SET pf_name = $2, pf_api_server = $3, pf_namespace = $4, pf_secret = $5
		WHERE pf_id = $1
		RETURNING pf_id, pf_name, COALESCE(pf_api_server, ''), COALESCE(pf_namespace, ''), COALESCE(pf_secret, '')
	`, p.Id, p.Name, p.ServerUri, p.Namespace, p.Secret)
}

func (s *platformStore) Delete(id string) (*api.Platform, error) {
	return s.queryOne(`
		DELETE FROM platforms
		WHERE pf_id = $1
		RETURNING pf_id, pf_name, COALESCE(pf_api_server, ''), COALESCE(pf_namespace, ''), COALESCE(pf_secret, '')
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
