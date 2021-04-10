package main

import (
	"fmt"
	"io"
	"os"
	"testing"

	artifactory "github.com/frohwerk/deputy-backend/internal/artifactory/client"
	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type repository struct{}

func (r *repository) Get(s string) (io.ReadCloser, error) {
	f := func(s string) string { return fmt.Sprintf("../../test/rthook/%v", s) }
	switch s {
	case "/some/path/hello-nodejs-1.0.tar.gz":
		return os.Open(f("test.tar.gz"))
	case "/some/path/test-1.0.zip":
		return os.Open(f("test.zip"))
	case "/some/path/test-1.0.jar":
		return os.Open(f("test.jar"))
	default:
		return nil, fmt.Errorf("missing")
	}
}

type store []database.File

func (s *store) CreateIfAbsent(f *database.File) (*database.File, error) {
	if i := s.index(f.Name, f.Digest); i > -1 {
		return &(*s)[i], nil
	}
	f.Id = uuid.NewString()
	*s = append(*s, *f)
	return f, nil
}

func (s *store) index(name, digest string) int {
	for i, v := range *s {
		if v.Name == name && v.Digest == digest {
			return i
		}
	}
	return -1
}

func (s *store) contains(name, digest string) assert.Comparison {
	return func() (success bool) {
		return s.index(name, digest) > -1
	}
}

func TestTarGzFile(t *testing.T) {
	store := &store{}
	h := EventHandler{Repository: &repository{}, FileCreater: store}
	h.OnArtifactDeployed(&artifactory.ArtifactInfo{Name: "hello-nodejs-1.0.tar.gz", Path: "/some/path/hello-nodejs-1.0.tar.gz"})
	assert.Condition(t, store.contains("hello-nodejs-1.0.tar.gz", "sha256:9d439c526ffb97b4d4dcbfd1d809d446001426a8d1ba666101c822f5b640afc8"))
	assert.Condition(t, store.contains("app/app.js", "sha256:76a7059dc31c6bec6d0597bc500a093d4d5d914c35f83dcf58703abf2e6c1fe6"))
	assert.Len(t, *store, 2)
}

func TestZip(t *testing.T) {
	store := &store{}
	h := EventHandler{Repository: &repository{}, FileCreater: store}
	h.OnArtifactDeployed(&artifactory.ArtifactInfo{Name: "test-1.0.zip", Path: "/some/path/test-1.0.zip"})
	// assert.Condition(t, store.contains("test-1.0.zip", "sha256:c0c49cbfb6e7f787117b3befda9992de781732549a6883ef7773ea037c866d63"))
	assert.Condition(t, store.contains("test-1.0.zip", "sha256:9491d551a065d79728024520dbe87972f6b5a5c2d94e05018f5ca05b69cfcda1"))
	assert.Condition(t, store.contains("app/app.js", "sha256:76a7059dc31c6bec6d0597bc500a093d4d5d914c35f83dcf58703abf2e6c1fe6"))
	assert.Condition(t, store.contains("test.txt", "sha256:a582e8c28249fe7d7990bfa0afebd2da9185a9f831d4215b4efec74f355b301a"))
	assert.Len(t, *store, 3)
}

func TestJar(t *testing.T) {
	store := &store{}
	h := EventHandler{Repository: &repository{}, FileCreater: store}
	h.OnArtifactDeployed(&artifactory.ArtifactInfo{Name: "test-1.0.jar", Path: "/some/path/test-1.0.jar"})
	assert.Condition(t, store.contains("test-1.0.jar", "sha256:baa9a16b0de61ba7de3823db6d18a9ca3d4cb05e4700b8ce723aab52b9c6bbe9"))
	assert.Condition(t, store.contains("META-INF/MANIFEST.MF", "sha256:1e719e8ac1ffaa167389430795aea28a575b84128cd3843c0ae89851cec3dec2"))
	assert.Condition(t, store.contains("META-INF/maven/de.frohwerk/hello-world/pom.properties", "sha256:ede0aabfea1f8783865b2bbd533c3c4ab1ccbc3a8a026defd117c9b305ebfeef"))
	assert.Condition(t, store.contains("META-INF/maven/de.frohwerk/hello-world/pom.xml", "sha256:2c1fa35b814aac113961527b9853f3e0edc0fa4dd44b9b879710120a6e421641"))
	assert.Condition(t, store.contains("de/frohwerk/HelloWorld$1.class", "sha256:aeaddd43e979059b7c952db89472d2f5ceeba690e34634207d78aff4c50e1b0a"))
	assert.Condition(t, store.contains("de/frohwerk/HelloWorld.class", "sha256:94d1a89bb6f0c6e3a6329bdc467b610e58a3d0135b86d1c88aad01c538c32e95"))
	assert.Len(t, *store, 6)
}
