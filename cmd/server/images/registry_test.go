package images_test

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/distribution/distribution/registry/api/errcode"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/frohwerk/deputy-backend/cmd/server/images"
	"github.com/stretchr/testify/assert"
)

const (
	Binary     = "application/octet-stream"
	ManifestV2 = "application/vnd.docker.distribution.manifest.v2+json"
)

type resource struct {
	location    string
	contentType string
}

type resources map[string]resource

func TestStuff(t *testing.T) {
	server := httptest.NewServer(mockImageRegistry(t, resources{
		"/myproject/jetty-hello-world/manifests/sha256:f1966dbfe1d5af2f0fe5779025368aa42883ba7a188a590f64b964e0fd01eeb3": resource{f("image/manifest.json"), ManifestV2},
		"/myproject/jetty-hello-world/blobs/sha256:5216c3c29ac4e62b214874a54a0125476beb1ee475d91a93444af351763c4629":     resource{f("image/layer.tar"), Binary},
	}))
	defer server.Close()

	r := images.RemoteRegistry{BaseUrl: server.URL}

	ctx := context.Background()
	m, err := r.Manifest(ctx, "172.30.1.1:5000/myproject/jetty-hello-world/sha256:f1966dbfe1d5af2f0fe5779025368aa42883ba7a188a590f64b964e0fd01eeb3")
	if err != nil {
		t.Errorf("Fetching the manifest failed: %s", err)
	}

	switch m := m.(type) {
	case *schema2.DeserializedManifest:
		assert.Equal(t, "sha256:0bec83099217b0784b60488337d97faf2d9c85d220248bd7a8f1e402b25a729c", m.Config.Digest.String())
		br, err := r.Blob(ctx, fmt.Sprintf("172.30.1.1:5000/myproject/jetty-hello-world/%s", m.Layers[len(m.Layers)-1].Digest))
		if err != nil {
			t.Errorf("Fetching the blob failed: %s", err)
		}
		tr := tar.NewReader(br)
		files := []string(nil)
		for th, err := tr.Next(); err != io.EOF; th, err = tr.Next() {
			if err != nil {
				t.Errorf("Reading the next tar header failed: %s", err)
			}
			files = append(files, th.Name)
		}
		assert.Equal(t, []string{}, files)
	default:
		t.Errorf("Expected type: schema2.DeserializedManifest v2\nActual type: %t", m)
	}

}

func f(name string) string {
	return fmt.Sprintf("../../../test/%s", name)
}

func mockImageRegistry(t *testing.T, resources map[string]resource) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		ref := strings.TrimPrefix(r.URL.String(), "/v2")
		fmt.Println("Incoming request", r.URL.String())
		if res, ok := resources[ref]; ok {
			buf, err := os.ReadFile(res.location)
			if err != nil {
				sendError(rw, http.StatusInternalServerError, 7, fmt.Sprintf("%s", err))
			} else {
				rw.Header().Set("Content-Type", res.contentType)
				rw.Write(buf)
			}
		} else {
			sendError(rw, http.StatusNotFound, 7, fmt.Sprintf("Object '%s' not found", ref))
		}
	}
}

func sendError(rw http.ResponseWriter, sc, ec int, m string) {
	e := errcode.Error{Code: errcode.ErrorCode(ec), Message: m}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(sc)
	if buf, err := json.Marshal(errcode.Errors{e}); err != nil {
		fmt.Printf("%s\n", err)
		rw.Write(nil)
	} else {
		rw.Write(buf)
	}
}
