package tar

import (
	"archive/tar"
	"strings"

	"github.com/frohwerk/deputy-backend/internal/fs"
)

func FromTarReader(name string, tr *tar.Reader) (*fs.Archive, error) {
	fsi, err := fs.FromIterator(newTarFsIterator(tr))
	if err != nil {
		return nil, err
	}
	return &fs.Archive{Name: name, FileSystemInfo: fsi}, nil
}

type tarFileSystem struct {
	tr *tar.Reader
}

func newTarFsIterator(tr *tar.Reader) fs.FileSystemIterator {
	return &tarFileSystem{tr}
}

func (tfs *tarFileSystem) Next() (*fs.FileSystemEntry, error) {
	for {
		h, err := tfs.tr.Next()
		switch {
		case err != nil:
			return nil, err
		case h.FileInfo().IsDir():
			continue
		default:
			return &fs.FileSystemEntry{Name: strings.TrimPrefix(h.Name, "./"), Reader: tfs.tr}, nil
		}
	}
}
