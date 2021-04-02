package database

import (
	"log"

	"github.com/frohwerk/deputy-backend/internal/util"
)

type Files []File

// An archive is a file containing other files (e.g. zip, tar, ...)
type Archive struct {
	File
	Files
}

func (f Files) Len() int {
	return len(f)
}

func (f Files) Less(i, j int) bool {
	return f[i].Name < f[j].Name
}

func (f Files) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

var _ ArchiveLookup = &fileStore{}

type ArchiveLookup interface {
	// Find archives containing the specified file
	FindByContent(f *File) ([]Archive, error)
}

func (s *fileStore) FindByContent(f *File) ([]Archive, error) {
	// TODO: Add second filter criteria: name (without path!)
	rows, err := s.db.Query(`
		SELECT DISTINCT a.file_id, a.file_digest, a.file_path
		  FROM files f
		  JOIN files a ON a.file_id = f.file_parent
		 WHERE f.file_digest = $1
		   AND a.file_parent is null
	`, f.Digest)
	if err != nil {
		return nil, err
	}
	defer util.Close(rows, log.Printf)
	archives := make([]Archive, 0)
	for ct := 0; rows.Next(); ct++ {
		archives = append(archives, Archive{})
		if err := rows.Scan(&archives[ct].Id, &archives[ct].Digest, &archives[ct].Name); err != nil {
			return nil, err
		}
		files, err := s.FindByParent(archives[ct].Id)
		if err != nil {
			return nil, err
		}
		archives[ct].Files = files
	}
	return archives, nil
}
