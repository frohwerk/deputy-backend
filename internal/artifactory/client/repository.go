package client

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Repository interface {
	Get(string) (io.ReadCloser, error)
}

type Artifactory struct {
	baseUri    string
	httpclient *http.Client
	Webhooks
}

func New(baseUri string) *Artifactory {
	return &Artifactory{baseUri: strings.TrimSuffix(baseUri, "/")}
}

func WithHttpClient(baseUri string, client *http.Client) *Artifactory {
	return &Artifactory{baseUri: strings.TrimSuffix(baseUri, "/"), httpclient: client}
}

func (rt *Artifactory) Get(uri string) (io.ReadCloser, error) {
	uri = strings.TrimPrefix(uri, "/")
	if rt.baseUri != "" {
		uri = fmt.Sprintf("%s/%s", rt.baseUri, uri)
	}
	resp, err := rt.client().Get(uri)
	switch {
	case err != nil:
		return nil, err
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("failed to download artifact: %v %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	default:
		return resp.Body, nil
	}
}

func (rt *Artifactory) client() *http.Client {
	if rt.httpclient == nil {
		return http.DefaultClient
	}
	return rt.httpclient
}
