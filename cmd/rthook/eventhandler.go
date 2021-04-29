package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"strings"

	"github.com/frohwerk/deputy-backend/internal/fs"
	tarfs "github.com/frohwerk/deputy-backend/internal/fs/tar"
	zipfs "github.com/frohwerk/deputy-backend/internal/fs/zip"
	"github.com/opencontainers/go-digest"

	artifactory "github.com/frohwerk/deputy-backend/internal/artifactory/client"
	"github.com/frohwerk/deputy-backend/internal/database"
)

type EventHandler struct {
	artifactory.Repository
	database.FileCreater
}

func (h *EventHandler) OnArtifactDeployed(i *artifactory.ArtifactInfo) error {
	r, err := h.Repository.Get(i.Path)
	if err != nil {
		return err
	}
	defer r.Close()
	archive, err := read(i.Name, r)
	if err != nil {
		return err
	}
	entity, err := h.CreateIfAbsent(&database.File{Digest: archive.Digest, Name: archive.Name})
	if err != nil {
		return err
	}
	for _, f := range archive.Files {
		if _, err := h.CreateIfAbsent(&database.File{Name: f.Name, Digest: f.Digest, Parent: entity.Id}); err != nil {
			return err
		}
	}
	log.Printf("Archive registered: %s", entity)
	return nil
}

func read(name string, r io.ReadCloser) (*fs.Archive, error) {
	switch {
	case strings.HasSuffix(name, ".jar"):
		fallthrough
	case strings.HasSuffix(name, ".zip"):
		return readZip(name, r)
	case strings.HasSuffix(name, ".tar.gz"):
		return readTarGz(name, r)
	default:
		dgst, err := digest.FromReader(r)
		if err != nil {
			return nil, err
		}
		return &fs.Archive{Name: name, FileSystemInfo: &fs.FileSystemInfo{Digest: dgst.String()}}, nil
		// return nil, fmt.Errorf("%s file format not supported yet\n", name)
	}
}

func readZip(name string, r io.ReadCloser) (*fs.Archive, error) {
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

func readTarGz(name string, r io.ReadCloser) (*fs.Archive, error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()
	return tarfs.FromTarReader(name, tar.NewReader(gzr))
}
