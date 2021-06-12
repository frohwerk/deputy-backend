package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/frohwerk/deputy-backend/internal/database/sqlstate"
	"github.com/frohwerk/deputy-backend/internal/util"
	"github.com/google/uuid"
)

type File struct {
	Id     string
	Name   string
	Digest string
	Parent string
}

type FileDigestFinder interface {
	FindByDigest(string) ([]File, error)
}

type FileLookup interface {
	Get(string) (*File, error)
	FindByParent(string) ([]File, error)
	FileDigestFinder
}

type FileCreater interface {
	// Create(f *File) (*File, error)
	CreateIfAbsent(f *File) (*File, error)
}

type FileStore interface {
	ArchiveLookup
	FileCreater
	FileLookup
}

type fileStore struct {
	db *sql.DB
}

func NewFileStore(db *sql.DB) *fileStore {
	return &fileStore{db}
}

func (s *fileStore) Create(f *File) (*File, error) {
	return s.selectFile(`
		INSERT INTO files (id, digest, path, parent_id)
		VALUES ($1, $2, $3, NULLIF($4, ''))
		RETURNING id, digest, path, COALESCE(parent_id, '')
	`, uuid.New().String(), f.Digest, f.Name, f.Parent)
}

func (s *fileStore) CreateIfAbsent(f *File) (*File, error) {
	c, err := s.Create(f)
	switch {
	case sqlstate.UniqueViolation(err):
		return s.FindByDigestAndPath(f.Digest, f.Name)
	case err != nil:
		return nil, err
	default:
		return c, nil
	}
}

func (s *fileStore) Get(id string) (*File, error) {
	return s.selectFile(`
		SELECT file_id, file_digest, file_path, COALESCE(file_parent, '')
		FROM files
		WHERE file_id = $1
	`, id)
}

func (s *fileStore) FindByDigestAndPath(digest, path string) (*File, error) {
	return s.selectFile(`
		SELECT file_id, file_digest, file_path, COALESCE(file_parent, '')
		FROM files
		WHERE file_digest = $1 AND file_path = $2
	`, digest, path)
}

func (s *fileStore) FindByDigest(digest string) ([]File, error) {
	return s.selectfiles(`
		SELECT file_id, file_digest, file_path, COALESCE(file_parent, '')
		FROM files
		WHERE file_digest = $1 AND file_parent is null
	`, digest)
}

func (s *fileStore) FindByParent(id string) ([]File, error) {
	return s.selectfiles(`
		SELECT file_id, file_digest, file_path, file_parent
		FROM files
		WHERE file_parent = $1
	`, id)
}

func (s *fileStore) selectFile(query string, args ...interface{}) (*File, error) {
	row := s.db.QueryRow(query, args...)
	f := &File{}
	if err := row.Scan(&f.Id, &f.Digest, &f.Name, &f.Parent); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *fileStore) selectfiles(query string, args ...interface{}) ([]File, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer util.Close(rows, log.Printf)
	files := make([]File, 0)
	for rows.Next() {
		f := File{}
		if err := rows.Scan(&f.Id, &f.Digest, &f.Name, &f.Parent); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

func (f *File) String() string {
	return fmt.Sprintf(`File{Id:"%s",Digest:"%s",Path:"%s",Parent:"%s"}`, f.Id, f.Digest, f.Name, f.Parent)
}
