package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	_ "crypto/sha256"

	"github.com/distribution/distribution/registry/client"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/reference"
	"github.com/frohwerk/deputy-backend/cmd/server/images"
	"github.com/stretchr/testify/assert"
)

func TestReferences(t *testing.T) {
	refs := []string{
		"node",
		"172.30.1.1:5000/myproject/jetty-hello-world@sha256:f1966dbfe1d5af2f0fe5779025368aa42883ba7a188a590f64b964e0fd01eeb3",
		"172.30.1.1:5000/myproject/jetty-hello-world:1.0",
		"myproject/jetty-hello-world@sha256:f1966dbfe1d5af2f0fe5779025368aa42883ba7a188a590f64b964e0fd01eeb3",
		"myproject/jetty-hello-world:1.0",
	}
	for _, s := range refs {
		ref, err := reference.Parse(s)
		if err != nil {
			t.Errorf("'%s': %s\n", s, err)
			continue
		}
		t.Logf("%T %s", ref, s)
		switch ref := ref.(type) {
		case reference.Named:
			t.Logf("Name: %s", ref.Name())
			t.Logf("Domain: %s", reference.Domain(ref))
			t.Logf("Path: %s", reference.Path(ref))

			if repo, err := client.NewRepository(ref, "http://ocrproxy-myproject.192.168.178.31.nip.io", http.DefaultTransport); err != nil {
				t.Errorf("failed to create new repository: %s", err)
			} else {
				if _, err := repo.Manifests(context.TODO()); err != nil {
					t.Errorf("failed to create distribution.ManifestService: %s", err)
				}
			}
		}
		switch ref := ref.(type) {
		case reference.Tagged:
			t.Logf("Tag: %s", ref.Tag())
		}
		switch ref := ref.(type) {
		case reference.Digested:
			t.Logf("Digest: %s", ref.Digest())
		}
	}
}

func ParseNamedReference(s string) (reference.Named, error) {
	ref, err := reference.ParseAnyReference(s)
	if err != nil {
		return nil, err
	}
	switch ref := ref.(type) {
	case reference.Named:
		return ref, nil
	default:
		return nil, fmt.Errorf("Unsupported reference type: %T", ref)
	}
}

func TestRegistry(t *testing.T) {
	refs := []string{
		// "myproject/node-hello-world:1.0.3",
		"172.30.1.1:5000/myproject/node-hello-world:1.0.3",
		"172.30.1.1:5000/myproject/jetty-hello-world@sha256:f1966dbfe1d5af2f0fe5779025368aa42883ba7a188a590f64b964e0fd01eeb3",
	}
	for _, s := range refs {
		ref, err := ParseNamedReference(s)
		if !assert.NoError(t, err) {
			continue
		}
		registry := &images.Registry{BaseUrl: "http://ocrproxy-myproject.192.168.178.31.nip.io"}
		repo, err := registry.Repository(ref)
		if !assert.NoError(t, err, "failed to create repository with reference '%s'", s) {
			continue
		}
		m, err := repo.Manifest(context.TODO(), ref)
		if !assert.NoErrorf(t, err, "failed to fetch distribution.Manifest: %s", err) {
			continue
		}
		files := []string(nil)
		switch m := m.(type) {
		case *schema2.DeserializedManifest:
			dgst := m.Layers[len(m.Layers)-1].Digest
			br, err := repo.Blob(context.TODO(), dgst)
			if !assert.NoErrorf(t, err, "failed to open reader for layer '%s': %s", dgst, err) {
				continue
			}
			defer br.Close()
			gzr, err := gzip.NewReader(br)
			if !assert.NoErrorf(t, err, "failed to open gzip.Reader for layer '%s': %s", dgst, err) {
				continue
			}
			defer gzr.Close()
			tr := tar.NewReader(gzr)
			for th, err := tr.Next(); err != io.EOF; th, err = tr.Next() {
				switch {
				case !assert.NoErrorf(t, err, "failed to read tar header: %s", err):
					break
				case th.FileInfo().IsDir():
					continue
				}
				files = append(files, th.Name)
			}
		default:
			assert.Fail(t, "expected type schema2.DeserializedManifest\nactual type %T", m)
		}
		fmt.Printf("files: %s\n", files)
		// TODO: Test the repository methods!
		// TODO: Create custom interface to decouple from docker distribution client
	}
}

func TestLookup(t *testing.T) {
	// ref, err := reference.ParseNamed("172.30.1.1:5000/myproject/node-hello-world@sha256:3bf137c335a2f7f9040eef6c2093abaa273135af0725fdeea5c4009a695d840f")
	ref, err := reference.ParseNamed("172.30.1.1:5000/myproject/node-hello-world:1.0.3")
	if err != nil {
		t.Errorf("Failed to parse reference: %s", err)
		return
	}

	// registry := images.NewRegistry("http://ocrproxy-myproject.192.168.178.31.nip.io")
	repoName, err := reference.WithName(reference.Path(ref))
	if err != nil {
		t.Errorf("Failed to create repository reference: %s", err)
		return
	}
	repo, err := client.NewRepository(repoName, "http://ocrproxy-myproject.192.168.178.31.nip.io", http.DefaultTransport)
	if err != nil {
		t.Errorf("Failed to parse reference: %s", err)
		return
	}
	tags := repo.Tags(context.TODO())
	manifests, err := repo.Manifests(context.TODO())
	if err != nil {
		t.Errorf("Failed to create distribution.ManifestService: %s", err)
		return
	}
	if tagged, ok := ref.(reference.Tagged); ok {
		d, err := tags.Get(context.TODO(), tagged.Tag())
		if err != nil {
			t.Errorf("Failed to lookup tag '%s': %s", ref, err)
			return
		}
		ref, err = reference.WithDigest(ref, d.Digest)
		if err != nil {
			t.Errorf("Failed to set digest for reference '%s': %s", ref, err)
			return
		}
	}
	if ref, ok := ref.(reference.Digested); ok {
		m, err := manifests.Get(context.TODO(), ref.Digest())
		if err != nil {
			t.Errorf("Failed to fetch Manifest: %s", err)
			return
		}
		if m, ok := m.(*schema2.DeserializedManifest); ok {
			t.Logf("Fetched manifest. Config reference: %s", m.Config.Digest)
		} else {
			t.Logf("Expected type *schema2.DeserializedManifest\nActual type: %T", m)
		}
	} else {
		t.Logf("Expected type reference.Digested\nActual type: %T", ref)
	}
}
