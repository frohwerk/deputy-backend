package database

import "database/sql"

type imageStore struct {
	db *sql.DB
}

type ImageLink struct {
	Id     string
	FileId string
}

type ImageLinker interface {
	AddLink(id, fileId string) (*ImageLink, error)
}

func NewImageStore(db *sql.DB) ImageLinker {
	return &imageStore{db}
}

func (s *imageStore) AddLink(id, fileId string) (*ImageLink, error) {
	return s.selectRow(`
		INSERT INTO images_artifacts (image_id, file_id)
		VALUES ($1, $2)
		RETURNING image_id, file_id
	`, id, fileId)
}

func (s *imageStore) selectRow(query string, args ...interface{}) (*ImageLink, error) {
	row := s.db.QueryRow(query, args...)
	i := &ImageLink{}
	if err := row.Scan(&i.Id, &i.FileId); err != nil {
		return nil, err
	}
	return i, nil
}
