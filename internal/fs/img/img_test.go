package img_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/distribution/distribution/registry/api/errcode"
	"github.com/frohwerk/deputy-backend/cmd/server/images"
	imgfs "github.com/frohwerk/deputy-backend/internal/fs/img"
	"github.com/stretchr/testify/assert"
)

var transport http.RoundTripper = &mockTransport{}

type mockTransport struct{}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	switch req.Method {
	case http.MethodHead:
		return t.Head(req)
	case http.MethodGet:
		return t.Get(req)
	default:
		return &http.Response{StatusCode: http.StatusMethodNotAllowed}, nil
	}
}

func (t *mockTransport) Head(req *http.Request) (*http.Response, error) {
	path := req.URL.Path
	fmt.Printf("Incoming: HEAD %s\n", path)
	switch path {
	case "/v2/myproject/my-image/manifests/1.2.3":
		h := make(http.Header)
		h.Add("Content-Length", "2097") // Does not matter as long as it's an unsigned integer
		h.Add("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
		h.Add("Docker-Content-Digest", "sha256:3bf137c335a2f7f9040eef6c2093abaa273135af0725fdeea5c4009a695d840f")
		return &http.Response{StatusCode: http.StatusOK, Header: h}, nil
	default:
		return t.Error(errcode.ErrorCodeUnknown, "requested resource does not exist")
	}
}

func (t *mockTransport) Get(req *http.Request) (*http.Response, error) {
	path := req.URL.Path
	fmt.Printf("Incoming: GET %s\n", path)
	switch path {
	case "/v2/myproject/my-image/manifests/sha256:3bf137c335a2f7f9040eef6c2093abaa273135af0725fdeea5c4009a695d840f":
		return t.File("fake-image/manifest.json", "application/vnd.docker.distribution.manifest.v2+json")
	case "/v2/myproject/my-image/blobs/sha256:714453066ddf1f2af9d71c34b65e0f6200a5183a57c18deed63bf7c77b599bbd":
		return t.File("fake-image/71445306.tar.gz", "application/vnd.docker.image.rootfs.diff.tar.gzip")
	case "/v2/myproject/my-image/blobs/sha256:3a25aa7cec0a27c7a87e3d08230cc69015698cbee90dcf3348460ab1351e08b5":
		return t.File("fake-image/3a25aa7c.tar.gz", "application/vnd.docker.image.rootfs.diff.tar.gzip")
	case "/v2/myproject/my-image/blobs/sha256:5216c3c29ac4e62b214874a54a0125476beb1ee475d91a93444af351763c4629":
		return t.File("fake-image/5216c3c2.tar.gz", "application/vnd.docker.image.rootfs.diff.tar.gzip")
	default:
		return t.Error(errcode.ErrorCodeUnknown, "requested resource does not exist")
	}
}

func (t *mockTransport) File(name, contentType string) (*http.Response, error) {
	wd, _ := os.Getwd()
	fmt.Printf("os.Getwd() = %s\n", wd)

	f, err := os.Open(fmt.Sprintf(`../../../../test/%s`, name))
	if err != nil {
		return t.Catch(err)
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		return t.Catch(err)
	}

	headers := make(http.Header)
	headers.Add("Content-Type", contentType)
	return &http.Response{StatusCode: http.StatusOK, Header: headers, Body: io.NopCloser(bytes.NewReader(buf))}, nil
}

func (t *mockTransport) Catch(err error) (*http.Response, error) {
	return t.Error(errcode.ErrorCodeUnavailable, fmt.Sprint(err))
}

func (t *mockTransport) Error(ec errcode.ErrorCode, msg string) (*http.Response, error) {
	body, err := json.Marshal(errcode.Error{Code: errcode.ErrorCodeUnknown, Message: "requested resource does not exist"})
	if err != nil {
		body = []byte(fmt.Sprintf(`{"Code":1,Message:"%s"}`, err))
	}
	return &http.Response{StatusCode: http.StatusMethodNotAllowed, Body: io.NopCloser(bytes.NewReader(body))}, nil
}

func TestFromImage(t *testing.T) {
	ref := "172.0.0.1:5000/myproject/my-image:1.2.3"
	registry := &images.RemoteRegistry{BaseUrl: "https://registry.server", Transport: transport}
	fs, err := imgfs.FromImage(ref, registry)
	if assert.NoError(t, err) && assert.NotNil(t, fs, "Either fs or err should have a non-nil value") {
		assert.Equal(t, ref, fs.Name)
		files := []string(nil)
		for _, f := range fs.Files {
			id := fmt.Sprintf("%s;%s", f.Name, f.Digest)
			fmt.Println(id)
			files = append(files, id)
		}
		assert.Contains(t, files, "/etc/debian_version;sha256:cce26cfeeb72d7a5c4b41df5bda474c75f7525783669dadb4b519efa79fded34")
		assert.Contains(t, files, "/etc/hostname;sha256:4796631793e89e4d6b5b2037536ee5aa85ed7a4168b139ecdaee6a4a55b03468")
		assert.Contains(t, files, "/etc/hosts;sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
		assert.Contains(t, files, "/app/stuff.txt;sha256:9652691c874495eab633b1082c6229189907e6d3bd6253bf1cdd3d92bacb4711")
		assert.Contains(t, files, "/app/app.js;sha256:17eb77dcb21a393822254cd1957ac4ab6e69de9d74bfa09aa45c0a6e73e900e4")
		assert.Equal(t, len(files), 5)
	}
}
