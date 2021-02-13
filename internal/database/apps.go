package database

import (
	"database/sql"
	"log"

	"github.com/google/uuid"

	"github.com/frohwerk/deputy-backend/internal/util"
)

type App struct {
	Id   string
	Name string
}

type AppStore interface {
	Get(string) (*App, error)
	Create(App) (*App, error)
	ListAll() ([]App, error)
}

type appStore struct {
	db *sql.DB
}

func NewAppStore(db *sql.DB) *appStore {
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
