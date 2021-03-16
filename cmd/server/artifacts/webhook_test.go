package artifacts

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	artifactory "github.com/frohwerk/deputy-backend/internal/artifactory/client"
	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	event, err := decode(strings.NewReader(`{
		"domain": "artifact",
		"event_type": "deployed",
		"data": {
			"name": "demo-0.0.1.jar",
			"path": "com/example/demo/0.0.1/demo-0.0.1.jar",
			"repo_key": "libs-release-local",
			"sha256": "56bf6ac8bc5319c64124512956859f54acfd86a9462cabaf9b687844d209be68",
			"size": 29390
		}
	}`))
	assert.NoError(t, err)
	assert.Equal(t, "artifact", event.Domain)
	assert.Equal(t, "deployed", event.EventType)
	assert.IsType(t, &artifactory.ArtifactInfo{}, event.Data)
	a := event.Data.(*artifactory.ArtifactInfo)
	assert.Equal(t, "demo-0.0.1.jar", a.Name)
	assert.Equal(t, "com/example/demo/0.0.1/demo-0.0.1.jar", a.Path)
	assert.Equal(t, "libs-release-local", a.Repo)
	assert.Equal(t, "56bf6ac8bc5319c64124512956859f54acfd86a9462cabaf9b687844d209be68", a.Sha256)
	assert.Equal(t, 29390, a.Size)
}

func TestOnArtifactDeployed(t *testing.T) {
	repo := make(mockRepository, 0)
	store := make(mockStore, 0)
	rec := httptest.NewRecorder()
	h := NewWebhookHandler(repo, &store)
	h.ServeHTTP(rec, &http.Request{
		Body: io.NopCloser(strings.NewReader(`{
			"domain": "artifact",
			"event_type": "deployed",
			"data": {
				"name": "demo-0.0.1.jar",
				"path": "com/example/demo/0.0.1/demo-0.0.1.jar",
				"repo_key": "libs-release-local",
				"sha256": "56bf6ac8bc5319c64124512956859f54acfd86a9462cabaf9b687844d209be68",
				"size": 29390
			}
		}`)),
	})
	// TODO Implement deterministic hashing of zip file contents and check result in mockStore. See filesystem.go in workshop
	assert.Len(t, store, 1)
	a := store[0]
	assert.Equal(t, "sha256:f253d29b7b5857cbf6544b3e88c7abde84cc5f012f4e539a0cbded1360f7acb9", a.Id)
	assert.Equal(t, "com/example/demo/0.0.1/demo-0.0.1.jar", a.Name)
}

type mockRepository []string

type mockStore []*database.Artifact

func (r mockRepository) Get(uri string) (io.ReadCloser, error) {
	if f, err := os.Open("../../../test/test-linux.zip"); err != nil {
		return nil, err
	} else {
		return f, nil
	}
}

func (s *mockStore) Create(id, name string) (*database.Artifact, error) {
	a := &database.Artifact{Id: id, Name: name}
	*s = append(*s, a)
	return a, nil
}

func (s *mockStore) CreateIfAbsent(id, name string) (*database.Artifact, error) {
	return s.Create(id, name)
}
