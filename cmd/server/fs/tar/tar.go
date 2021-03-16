package tar

import (
	"archive/tar"
	"strings"

	"github.com/frohwerk/deputy-backend/cmd/server/fs"
)

func FromTarReader(name string, tr *tar.Reader) (*fs.FileSystemDigests, error) {
	return fs.FromFilesystem(name, newTarFs(tr))
}

type tarFileSystem struct {
	tr *tar.Reader
}

func newTarFs(tr *tar.Reader) fs.FileSystem {
	return &tarFileSystem{tr}
}

func (tfs *tarFileSystem) Next() (*fs.FileSystemEntry, error) {
	h, err := tfs.tr.Next()
	if err != nil {
		return nil, err
	}
	return &fs.FileSystemEntry{
		Name:   strings.TrimPrefix(h.Name, "./"),
		Reader: tfs.tr,
	}, nil
}
