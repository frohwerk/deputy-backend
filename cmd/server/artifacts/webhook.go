package artifacts

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"

	artifactory "github.com/frohwerk/deputy-backend/internal/artifactory/client"
	"github.com/frohwerk/deputy-backend/internal/database"
	digest "github.com/opencontainers/go-digest"
)

type webhookHandler struct {
	repo  artifactory.Repository
	store database.ArtifactStore
}

func NewWebhookHandler(r artifactory.Repository, s database.ArtifactStore) http.Handler {
	return &webhookHandler{repo: r, store: s}
}

func (h *webhookHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	event, err := decode(r.Body)

	if err != nil {
		writeResponse(rw, http.StatusBadRequest, "Error decoding request message: %s", err)
	}

	switch object := event.Data.(type) {
	case *artifactory.ArtifactInfo:
		if err := h.onArtifactDeployed(object); err != nil {
			writeResponse(rw, http.StatusInternalServerError, "Error storing artifact: %s", err)
		}
	default:
		writeResponse(rw, http.StatusBadRequest, "Unsupported domain: '%v'", event.Domain)
	}
}

func (h *webhookHandler) onArtifactDeployed(i *artifactory.ArtifactInfo) error {
	r, err := h.repo.Get(i.Path)
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

	fn := make([]string, len(zr.File))
	fs := make(map[string]string)
	for i, f := range zr.File {
		fmt.Printf("ZIP-Entry: %s\n", f.Name)
		r, err := f.Open()
		if err != nil {
			return err
		}

		dgst, err := digest.FromReader(r)
		if err != nil {
			return err
		}

		fn[i] = f.Name
		fs[f.Name] = dgst.Encoded()
	}

	sort.Strings(fn)
	sb := strings.Builder{}
	for _, name := range fn {
		sb.WriteString(fmt.Sprintf("%s;%s\n", name, fs[name]))
	}
	dgst := digest.FromString(sb.String()).String()
	sb.Reset()

	fmt.Printf("Archive-Ref: %s\n", dgst)

	// TODO: Download artifact, create unique hash of the file system contained in the archive
	if _, err := h.store.Create(dgst, i.Path); err != nil {
		return err
	}

	return nil
}

func decode(r io.Reader) (*artifactory.ArtifactEvent, error) {
	event := new(artifactory.ArtifactEvent)
	object := new(json.RawMessage)
	event.Data = object
	if err := json.NewDecoder(r).Decode(event); err != nil {
		return nil, err
	}

	switch event.Domain {
	case "artifact":
		art := new(artifactory.ArtifactInfo)
		if err := json.Unmarshal(*object, art); err != nil {
			return nil, err
		}
		event.Data = art
	default:
		return nil, fmt.Errorf("Unsupported domain: '%v'", event.Domain)
	}

	return event, nil
}

func writeResponse(rw http.ResponseWriter, status int, msg string, args ...interface{}) {
	rw.WriteHeader(status)
	if _, err := rw.Write([]byte(fmt.Sprintf(msg, args...))); err != nil {
		log.Printf("Error sending response message: %v", err)
	}
}
