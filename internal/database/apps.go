package database

import (
	"context"
	"database/sql"
	"log"

	"github.com/google/uuid"

	"github.com/frohwerk/deputy-backend/internal/util"
)

type App struct {
	Id   string
	Name string
}

type AppUpdater interface {
	Update(*App) (*App, error)
}

type AppDeleter interface {
	Delete(string) (*App, error)
}

type AppStore interface {
	ListAll() ([]App, error)
	Create(App) (*App, error)
	Get(string) (*App, error)
	AppUpdater
	AppDeleter

	UpdateComponents(ctx context.Context, id string, components []string) error
}

type appStore struct {
	db *sql.DB
}

func NewAppStore(db *sql.DB) AppStore {
	return &appStore{db}
}

func (s *appStore) Get(id string) (*App, error) {
	row := s.db.QueryRow(`SELECT * FROM apps WHERE id = $1`, id)
	a := &App{}
	if err := row.Scan(&a.Id, &a.Name); err != nil {
		log.Printf("%t: %v\n", err, err)
		return nil, wrap(err)
	}
	return a, nil
}

func (s *appStore) ListAll() ([]App, error) {
	rows, err := s.db.Query(`SELECT * FROM apps`)
	if err != nil {
		return nil, err
	}
	defer util.Close(rows, log.Printf)

	apps := make([]App, 0)
	for rows.Next() {
		a := App{}
		if err := rows.Scan(&a.Id, &a.Name); err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}

	return apps, nil
}

func (s *appStore) Create(app App) (*App, error) {
	row := s.db.QueryRow(`INSERT INTO apps (id, name) VALUES($1, $2) RETURNING id, name`, string(uuid.NewString()), app.Name)
	a := new(App)
	if err := row.Scan(&a.Id, &a.Name); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *appStore) Update(app *App) (*App, error) {
	return s.queryOne(`UPDATE apps SET name = $2 WHERE id = $1 RETURNING id, name`, app.Id, app.Name)
}

func (s *appStore) Delete(id string) (*App, error) {
	return s.queryOne(`
		DELETE FROM apps
		WHERE id = $1
		RETURNING id, name
	`, id)
}

// TODO: Patching the assignments instead of recreating is probably more efficient
func (s *appStore) UpdateComponents(ctx context.Context, id string, components []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	delete, err := tx.Prepare(`DELETE FROM apps_components WHERE app_id = $1`)
	if err != nil {
		return err
	}

	if _, err := delete.Exec(id); err != nil {
		return tryRollback(tx, err)
	}

	insert, err := tx.Prepare(`INSERT INTO apps_components (app_id, component_id) VALUES($1, $2)`)
	if err != nil {
		return tryRollback(tx, err)
	}

	for _, componentId := range components {
		if _, err := insert.Exec(id, componentId); err != nil {
			return tryRollback(tx, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func tryRollback(tx *sql.Tx, err error) error {
	if err := tx.Rollback(); err != nil {
		return err
	}
	return err
}

func (s *appStore) queryOne(query string, args ...interface{}) (*App, error) {
	return scanApp(s.db.QueryRow(query, args...))
}

func scanApp(s scanner) (*App, error) {
	app := new(App)
	if err := s.Scan(&app.Id, &app.Name); err != nil {
		return nil, err
	}
	return app, nil
}
