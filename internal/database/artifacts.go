package database

import (
	"database/sql"
	"fmt"

	"github.com/frohwerk/deputy-backend/internal/database/sqlstate"
)

type Artifact struct {
	Id   string
	Name string
}

type ArtifactStore interface {
	Create(id, name string) (*Artifact, error)
	CreateIfAbsent(id, name string) (*Artifact, error)
}

type artifactStore struct {
	db *sql.DB
}

func NewArtifactStore(db *sql.DB) ArtifactStore {
	return &artifactStore{db}
}

func (s *artifactStore) Create(id, name string) (*Artifact, error) {
	return s.selectArtifact(`
		INSERT INTO artifacts (artifact_id, artifact_name)
		VALUES ($1, $2)
		RETURNING artifact_id, artifact_name
	`, id, name)
}

func (s *artifactStore) CreateIfAbsent(id, name string) (*Artifact, error) {
	c, err := s.Create(id, name)
	switch {
	case sqlstate.UniqueViolation(err):
		return &Artifact{Id: id, Name: name}, nil
	case err != nil:
		return nil, err
	default:
		return c, nil
	}
}

func (s *artifactStore) Get(id, name string) (*Artifact, error) {
	return s.selectArtifact(`
		SELECT artifact_id, artifact_name
		FROM artifacts
		WHERE artifact_id = $1 AND artifact_name = $2
	`, id, name)
}

func (s *artifactStore) selectArtifact(query string, args ...interface{}) (*Artifact, error) {
	row := s.db.QueryRow(query, args...)
	a := &Artifact{}
	if err := row.Scan(&a.Id, &a.Name); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *Artifact) String() string {
	return fmt.Sprintf(`Artifact{Id:"%s",Name:"%s"}`, a.Id, a.Name)
}
