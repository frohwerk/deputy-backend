package matcher_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/reference"
	"github.com/frohwerk/deputy-backend/cmd/imgmatch/matcher"
	"github.com/frohwerk/deputy-backend/cmd/server/images"
	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/google/uuid"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
)

var files []database.File

func init() {
	files = []database.File{
		{Id: "parent-id", Name: "app-1.0.tar.gz", Digest: "sha256:49dca07c575c4ea9e9a2cbbb550ba72ad93b3207086fa4f1cd17c03378d1b5c5", Parent: ""},
		{Id: uuid.NewString(), Name: "stuff.txt", Digest: "sha256:9652691c874495eab633b1082c6229189907e6d3bd6253bf1cdd3d92bacb4711", Parent: "parent-id"},
		{Id: uuid.NewString(), Name: "app.js", Digest: "sha256:17eb77dcb21a393822254cd1957ac4ab6e69de9d74bfa09aa45c0a6e73e900e4", Parent: "parent-id"},
	}
}

type mock struct{}

func (m *mock) findById(id string) (*database.File, error) {
	for _, f := range files {
		if f.Id == id {
			return &f, nil
		}
	}
	return nil, fmt.Errorf("file %s not found", id)
}

func (m *mock) FindByDigest(dgst string) ([]database.File, error) {
	var result []database.File
	for _, f := range files {
		if f.Parent == "" && f.Digest == dgst {
			result = append(result, f)
		}
	}
	return result, nil
}

func (m *mock) FindByContent(criteria *database.File) ([]database.Archive, error) {
	var result []database.Archive
	for _, f := range files {
		if f.Parent != "" && strings.HasSuffix(criteria.Name, f.Name) && f.Digest == criteria.Digest {
			parent, err := m.findById(f.Parent)
			if err != nil {
				return nil, err
			}
			archive := database.Archive{File: *parent}
			for _, f := range files {
				if f.Parent == archive.Id {
					archive.Files = append(archive.Files, f)
				}
			}
			result = append(result, archive)
		}
	}
	return result, nil
}

func (m *mock) Repository(named reference.Named) (images.Repository, error) {
	return m, nil
}

func (*mock) Manifest(ctx context.Context, ref reference.Reference, options ...distribution.ManifestServiceOption) (distribution.Manifest, error) {
	buf, err := os.ReadFile("../../../test/data/fake-image/manifest.json")
	if err != nil {
		return nil, err
	}
	m := schema2.DeserializedManifest{}
	if err := m.UnmarshalJSON(buf); err != nil {
		return nil, err
	}
	return &m, nil
}

func (m *mock) Blob(ctx context.Context, dgst digest.Digest) (io.ReadSeekCloser, error) {
	return os.Open(fmt.Sprintf("../../../test/data/fake-image/%s.tar.gz", dgst.Encoded()[:8]))
}

func TestMatch(t *testing.T) {
	mock := &mock{}
	matcher := matcher.New(mock, mock, mock)
	m, err := matcher.Match("172.30.1.1:5000/example/image:1.1")
	if assert.NoError(t, err) && assert.Len(t, m, 1) {
		assert.Equal(t, m[0].Id, "parent-id")
		assert.Equal(t, m[0].Name, "app-1.0.tar.gz")
		assert.Equal(t, m[0].Digest, "sha256:49dca07c575c4ea9e9a2cbbb550ba72ad93b3207086fa4f1cd17c03378d1b5c5")
		assert.Equal(t, m[0].Parent, "")
	}
	// TODO: Verify, that matcher has linked image and archive in the database (images_files table?)
	// HINT: It does not currently ;)
}
