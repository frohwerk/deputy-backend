package artifacts

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/frohwerk/deputy-backend/cmd/server/fs"
	tarfs "github.com/frohwerk/deputy-backend/cmd/server/fs/tar"
	zipfs "github.com/frohwerk/deputy-backend/cmd/server/fs/zip"

	artifactory "github.com/frohwerk/deputy-backend/internal/artifactory/client"
	"github.com/frohwerk/deputy-backend/internal/database"
)

type EventHandler struct {
	artifactory.Repository
	database.FileStore
}

func (h *EventHandler) OnArtifactDeployed(i *artifactory.ArtifactInfo) error {
	r, err := h.Repository.Get(i.Path)
	if err != nil {
		return err
	}
	defer r.Close()
	fsd, err := read(i.Name, r)
	if err != nil {
		return err
	}
	archive, err := h.CreateIfAbsent(&database.File{Digest: fsd.Digest, Name: fsd.Name})
	if err != nil {
		return err
	}
	for name, dgst := range fsd.FileDigests {
		if _, err := h.CreateIfAbsent(&database.File{Digest: dgst, Name: name, Parent: archive.Id}); err != nil {
			return err
		}
	}
	log.Printf("Archive registered: %s", archive)
	return nil
}

func read(name string, r io.ReadCloser) (*fs.FileSystemInfo, error) {
	switch {
	case strings.HasSuffix(name, ".zip"):
		return readZip(name, r)
	case strings.HasSuffix(name, ".tar.gz"):
		return readTarGz(name, r)
	default:
		return nil, fmt.Errorf("%s file format not supported yet\n", name)
	}
}

func readZip(name string, r io.ReadCloser) (*fs.FileSystemInfo, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	zr, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return nil, err
	}
	return zipfs.FromZipReader(name, zr)
}

func readTarGz(name string, r io.ReadCloser) (*fs.FileSystemInfo, error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()
	return tarfs.FromTarReader(name, tar.NewReader(gzr))
}
