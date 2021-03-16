package registry

import (
	"net/http"
	"net/url"

	"github.com/distribution/distribution/registry/client"
	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
)

type Registry interface {
	Repo(name string) (distribution.Repository, error)
	SetTransport(t http.RoundTripper) Registry
}

type registry struct {
	baseURL   string
	transport http.RoundTripper
}

func New(baseURL string) (Registry, error) {
	if _, err := url.Parse(baseURL); err != nil {
		return nil, err
	}
	return &registry{
		baseURL:   baseURL,
		transport: http.DefaultTransport,
	}, nil
}

func (r *registry) Repo(name string) (distribution.Repository, error) {
	ref, err := reference.WithName(name)
	if err != nil {
		return nil, err
	}
	return client.NewRepository(ref, r.baseURL, r.transport)
}

func (r *registry) SetTransport(t http.RoundTripper) Registry {
	r.transport = t
	return r
}
