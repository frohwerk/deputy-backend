package database

import (
	"database/sql"
	"log"
	"strings"

	"github.com/frohwerk/deputy-backend/internal/util"
)

type Env struct {
	Id   string
	Name string
}

type EnvCreator interface {
	Create(name string) (*Env, error)
}

type EnvDeleter interface {
	Delete(id string) (*Env, error)
}

type EnvGetter interface {
	Get(id string) (*Env, error)
}

type EnvUpdater interface {
	Update(env *Env) (*Env, error)
}

type EnvLister interface {
	List() ([]Env, error)
}

type EnvFinder interface {
	FindByName(name string) (*Env, error)
}

type EnvStore interface {
	EnvCreator
	EnvGetter
	EnvLister
	EnvFinder
	EnvDeleter
	EnvUpdater
}

type envStore struct {
	*sql.DB
}

func NewEnvStore(db *sql.DB) EnvStore {
	return &envStore{db}
}

func (s *envStore) Create(name string) (*Env, error) {
	return s.queryOne(`
		INSERT INTO envs (env_name) VALUES(NULLIF($1, ''))
		RETURNING env_id, env_name
	`, name)
}

func (s *envStore) Update(env *Env) (*Env, error) {
	return s.queryOne(`
		UPDATE envs SET
		env_name = NULLIF($2, '')
		WHERE env_id = $1
		RETURNING env_id, env_name
	`, env.Id, env.Name)
}

func (s *envStore) List() ([]Env, error) {
	return s.queryAll(`
		SELECT env_id, env_name
		FROM envs
	`)
}

func (s *envStore) Get(id string) (*Env, error) {
	return s.queryOne(`
		SELECT env_id, env_name
		FROM envs
		WHERE env_id = $1
	`, id)
}

func (s *envStore) FindByName(name string) (*Env, error) {
	return s.queryOne(`
		SELECT env_id, env_name
		FROM envs
		WHERE lower(env_name) = $1
	`, strings.ToLower(name))
}

func (s *envStore) Delete(id string) (*Env, error) {
	return s.queryOne(`
		DELETE FROM envs
		WHERE env_id = $1
		RETURNING env_id, env_name
	`, id)
}

func (s *envStore) queryOne(query string, args ...interface{}) (*Env, error) {
	return scanEnv(s.DB.QueryRow(query, args...))
}

func (s *envStore) queryAll(query string, args ...interface{}) ([]Env, error) {
	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer util.Close(rows, log.Printf)
	envs := make([]Env, 0)
	for rows.Next() {
		if env, err := scanEnv(rows); err != nil {
			return nil, err
		} else {
			envs = append(envs, *env)
		}
	}
	return envs, nil
}

func (s *envStore) exec(query string, args ...interface{}) error {
	_, err := s.DB.Exec(query, args...)
	if err != nil {
		return err
	}
	return nil

}

func scanEnv(s scanner) (*Env, error) {
	e := Env{}
	if err := s.Scan(&e.Id, &e.Name); err != nil {
		return nil, err
	}
	return &e, nil
}
