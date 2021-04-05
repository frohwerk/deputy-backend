package images

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"
)

type repository struct {
	distribution.Repository
}

type ManifestLookup interface {
	// Look up a manifest using the specified reference
	Manifest(ctx context.Context, ref reference.Reference, options ...distribution.ManifestServiceOption) (distribution.Manifest, error)
}

type BlobLookup interface {
	// Open an io.ReadSeekCloser for the specified digest
	Blob(ctx context.Context, dgst digest.Digest) (io.ReadSeekCloser, error)
}

type Repository interface {
	ManifestLookup
	BlobLookup
}

func (r *repository) Manifest(ctx context.Context, ref reference.Reference, options ...distribution.ManifestServiceOption) (distribution.Manifest, error) {
	ms, err := r.Repository.Manifests(ctx, options...)
	if err != nil {
		return nil, err
	}
	dgst, err := r.digest(ctx, ref)
	if err != nil {
		return nil, err
	}
	return ms.Get(ctx, dgst)
}

func (r *repository) Blob(ctx context.Context, dgst digest.Digest) (io.ReadSeekCloser, error) {
	return r.Repository.Blobs(ctx).Open(ctx, dgst)
}

func (r *repository) digest(ctx context.Context, ref reference.Reference) (digest.Digest, error) {
	switch ref := ref.(type) {
	case reference.Digested:
		return ref.Digest(), nil
	case reference.Tagged:
		desc, err := r.Repository.Tags(ctx).Get(ctx, ref.Tag())
		if err != nil {
			return "", err
		}
		return desc.Digest, nil
	default:
		return "", fmt.Errorf("reference type must be either reference.Digested or reference.Tagged. Actual type is %T", ref)
	}
}
