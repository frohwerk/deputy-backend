package client_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	artifactory "github.com/frohwerk/deputy-backend/internal/artifactory/client"
	"github.com/stretchr/testify/assert"
)

var mux *http.ServeMux
var server *httptest.Server

type mockHandler struct{}

func TestMain(m *testing.M) {
	mux = http.NewServeMux()
	server = httptest.NewTLSServer(mux)
	defer server.Close()
	os.Exit(m.Run())
}

func TestStuff(t *testing.T) {
	rt := artifactory.WithHttpClient(server.URL, server.Client())

	infos := make([]*artifactory.ArtifactInfo, 0)
	rt.OnArtifactDeployed(func(i *artifactory.ArtifactInfo) error {
		infos = append(infos, i)
		return nil
	})

	mux.HandleFunc("/webhooks/artifactory", rt.WebhookHandler)
	mux.HandleFunc("/com/example/demo/0.0.1/demo-0.0.1.jar", fromReader(t, bytes.NewBufferString("Content of artifact...")))

	uri := fmt.Sprintf("%s%s", server.URL, "/webhooks/artifactory")
	resp, err := server.Client().Post(uri, "application/json", strings.NewReader(`{
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
	assert.NoError(t, err, "Webhook invocation failed")

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "Reading response body failed")
	t.Log(string(body))

	assert.NoError(t, err, "Webhook invocation failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expecting http status 200 (OK)")
	assert.Len(t, infos, 1, "expecting exactly one ArtifactInfo")

	r, err := rt.Get(infos[0].Path)
	assert.NoError(t, err, "Failed to get io.ReadCloser for artifact")
	buf, err := io.ReadAll(r)
	assert.NoError(t, err, "Failed to read artifact")
	assert.Equal(t, "Content of artifact...", string(buf))
}

func fromReader(t *testing.T, src io.Reader) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		if _, err := io.Copy(rw, src); err != nil {
			t.Errorf("Failed to write response message: %s", err)
		}
	}
}

func (s *mockHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	log.Printf("path: %s", path)
	switch path {
	case "/com/example/demo/0.0.1/demo-0.0.1.jar":
		rw.WriteHeader(http.StatusOK)
		if _, err := rw.Write([]byte("Content of artifact...")); err != nil {
			log.Printf("Failed to write response message: %s", err)
		}
	default:
		rw.WriteHeader(http.StatusNotFound)
		if _, err := http.NoBody.WriteTo(rw); err != nil {
			log.Printf("Failed to write response message: %s", err)
		}
	}
}
