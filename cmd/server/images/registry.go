package images

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	_ "github.com/docker/distribution/manifest/schema2"

	"github.com/distribution/distribution/registry/client"
	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"
)

var _ Registry = &RemoteRegistry{}

type Registry interface {
	// Create repository instance  for a named repository
	Repository(named reference.Named) (Repository, error)
}

// A client abstraction for a docker image registry.
type RemoteRegistry struct {
	BaseUrl   string
	Transport http.RoundTripper
}

func (r *RemoteRegistry) Repository(ref reference.Named) (Repository, error) {
	path := reference.Path(ref)
	repo, err := reference.WithName(path)
	if err != nil {
		return nil, err
	}
	client, err := client.NewRepository(repo, r.BaseUrl, r.transport())
	if err != nil {
		return nil, err
	}
	return &repository{client}, nil
}

func (r *RemoteRegistry) Manifest(ctx context.Context, s string) (distribution.Manifest, error) {
	i := strings.Index(s, "/")
	j := strings.LastIndex(s, "/")

	repo, err := reference.WithName(s[i+1 : j])
	if err != nil {
		return nil, err
	}

	ref, err := reference.WithDigest(repo, digest.FromString(s[j+1:]))
	if err != nil {
		return nil, err
	}

	switch ref := ref.(type) {
	case reference.Canonical:
		repo, err := client.NewRepository(repo, r.BaseUrl, r.transport())
		if err != nil {
			return nil, err
		}
		ms, err := repo.Manifests(ctx)
		if err != nil {
			return nil, err
		}
		return ms.Get(ctx, ref.Digest())
	}
	// switch ref := ref.(type) {
	// case *schema2.DeserializedManifest:
	// 	return nil, nil
	// }
	return nil, fmt.Errorf("Not implemented yet")
}

func (r *RemoteRegistry) Blob(ctx context.Context, s string) (io.ReadSeekCloser, error) {
	i := strings.LastIndex(s, "/")

	repo, err := reference.WithName(s[:i])
	if err != nil {
		return nil, err
	}

	ref, err := reference.WithDigest(repo, digest.FromString(s[:i+1]))
	if err != nil {
		return nil, err
	}

	switch ref := ref.(type) {
	case reference.Canonical:
		repo, err := client.NewRepository(repo, r.BaseUrl, r.transport())
		if err != nil {
			return nil, err
		}
		bs := repo.Blobs(ctx)
		return bs.Open(ctx, ref.Digest())
	}
	return nil, fmt.Errorf("Not implemented yet")
}

func (r *RemoteRegistry) transport() http.RoundTripper {
	if r.Transport == nil {
		return http.DefaultTransport
	}
	return r.Transport
}
