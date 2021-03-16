package filesystem

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/opencontainers/go-digest"
)

type FileDescriptor struct {
	Path   string
	Name   string
	Digest string
}

type FileDigests map[string][]string

// type FileSystem struct {
// 	// An identical file can exist multiple times in a filesystem
// 	files map[string][]string
// 	// Each file should have exactly one digest
// 	digests map[string]string
// }

func FromArchive(filename string) (FileDigests, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	fs := make(FileDigests)
	tr := tar.NewReader(gzr)
	for h, err := tr.Next(); err != io.EOF; h, err = tr.Next() {
		d, err := digest.FromReader(tr)
		if err != nil {
			log.Fatalf("ERROR Failed to create digest from reader: %v\n", err)
		}
		name := strings.TrimPrefix(h.Name, "./")
		fs[d.Encoded()] = []string{name}
	}
	return fs, nil
}

func FromImage(r distribution.Repository, dgst digest.Digest) (FileDigests, error) {
	ctx := context.Background()

	ms, err := r.Manifests(ctx)
	if err != nil {
		log.Printf("ERROR Failed to create distribution.ManifestService: %s\n", err)
		return nil, err
	}

	mediaTypes := distribution.WithManifestMediaTypes([]string{
		// "application/vnd.oci.image.index.v1+json",
		// "application/vnd.docker.distribution.manifest.list.v2+json",
		// "application/vnd.oci.image.manifest.v1+json",
		"application/vnd.docker.distribution.manifest.v2+json",
		// "application/vnd.docker.distribution.manifest.v1+json",
		// "application/vnd.docker.distribution.manifest.v1+prettyjws",
	})
	m, err := ms.Get(ctx, dgst, mediaTypes)
	if err != nil {
		log.Printf("ERROR Failed to fetch manifest: %s\n", err)
		return nil, err
	}

	switch m := m.(type) {
	case *schema2.DeserializedManifest:
		return fromManifestV2(ctx, r, m)
	}
	return make(FileDigests), nil
}

func fromManifestV2(ctx context.Context, r distribution.Repository, m *schema2.DeserializedManifest) (FileDigests, error) {
	fs := make(FileDigests)
	log.Printf("DEBUG config.digest: %s\n", m.Config.Digest)
	if len(m.Layers) < 1 {
		return nil, errors.New("Manifest has no file system layers")
	}

	for _, layer := range m.Layers {
		log.Printf("DEBUG Layer %s\n", layer.Digest)
		buf, err := r.Blobs(ctx).Get(ctx, layer.Digest)
		if err != nil {
			log.Printf("ERROR Failed to fetch config blob: %s\n", err)
			return nil, err
		}

		log.Printf("TRACE Start processing layer filesystem: %s\n", time.Now().Format(time.RFC3339))
		gzr, err := gzip.NewReader(bytes.NewReader(buf))
		if err != nil {
			log.Printf("ERROR Failed to create gzip.Reader: %s\n", err)
			return nil, err
		}

		tr := tar.NewReader(gzr)
		for h, err := tr.Next(); err != io.EOF; h, err = tr.Next() {
			if err != nil {
				log.Printf("ERROR Failed to read next tar.Header: %s\n", err)
				return nil, err
			}
			// IMPORTANT TODO: Handle whiteouts!
			// OPTIMIZATION: Exclude certain directories from (first) scan
			if !h.FileInfo().IsDir() {
				d, err := digest.FromReader(tr)
				if err != nil {
					log.Printf("ERROR Failed to create digest from reader: %v\n", err)
					return nil, err
				}
				name := fmt.Sprintf("/%s", strings.TrimPrefix(h.Name, "./"))
				fs[d.Encoded()] = []string{name}
			}
		}
		log.Printf("TRACE End processing layer filesystem: %s\n", time.Now().Format(time.RFC3339))
	}
	return fs, nil
}

func (fs FileDigests) Contains(other FileDigests) bool {
	// Predicates:
	// - All hashes are found in the outer file system
	// - All matches share a common prefix in the outer file system
	counters := make(map[string]uint)
	for hash, ipaths := range other {
		if opaths, ok := fs[hash]; ok {
			log.Printf("DEBUG this <=> other: %s = %s\n", opaths, ipaths)
			for _, ipath := range ipaths {
				for _, opath := range opaths {
					prefix := strings.TrimSuffix(opath, ipath)
					if c, ok := counters[prefix]; ok {
						counters[prefix] = c + 1
					} else {
						counters[prefix] = 1
					}
				}
			}
		} else {
			log.Printf("DEBUG File %s (hash: %s) not found in file system\n", ipaths, hash)
			return false
		}
	}
	l := uint(len(other))
	log.Printf("DEBUG Number of expected matches: %v", l)
	for p, i := range counters {
		log.Printf("DEBUG Number of actual matches for path '%s': %v", p, i)
	}
	for p, i := range counters {
		if i == l {
			log.Printf("INFO  Found inner file system at path: %s\n", p)
			return true
		}
	}
	log.Print("INFO  Did NOT find inner file\n")
	return false
}
