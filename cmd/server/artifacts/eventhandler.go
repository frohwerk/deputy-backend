package artifacts

import (
	"archive/zip"
	"bytes"
	"io"
	"log"
	"strings"

	zipfs "github.com/frohwerk/deputy-backend/cmd/server/fs/zip"

	artifactory "github.com/frohwerk/deputy-backend/internal/artifactory/client"
	"github.com/frohwerk/deputy-backend/internal/database"
)

type EventHandler struct {
	artifactory.Repository
	database.ArtifactStore
}

func (h *EventHandler) OnArtifactDeployed(i *artifactory.ArtifactInfo) error {
	if !strings.HasSuffix(i.Name, ".jar") {
		log.Printf("%s file format not supported yet\n", i.Name)
		return nil
	}
	r, err := h.Repository.Get(i.Path)
	if err != nil {
		return err
	}
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return err
	}
	fsd, err := zipfs.FromZipReader(i.Name, zr)
	if err != nil {
		return err
	}
	a, err := h.ArtifactStore.CreateIfAbsent(fsd.Digest, fsd.Name)
	if err != nil {
		return err
	}
	for name, dgst := range fsd.FileDigests {
		if _, err := h.ArtifactStore.CreateIfAbsent(dgst, name); err != nil {
			return err
		}
	}
	log.Printf("Artifact created: %s", a)
	return nil
}
