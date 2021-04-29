package img

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/frohwerk/deputy-backend/cmd/imgmatch/predicates"
	"github.com/frohwerk/deputy-backend/cmd/server/images"
	"github.com/frohwerk/deputy-backend/internal/fs"
	"github.com/opencontainers/go-digest"
)

// Read an image filesystem from an image registry, ignore files matched by one of the ignores predicates.Predicates
func FromImage(ref string, registry images.Registry, ignores ...predicates.Predicate) (*fs.Archive, error) {
	archive := &fs.Archive{Name: ref, FileSystemInfo: &fs.FileSystemInfo{Files: make(fs.FileSlice, 0)}}
	named, err := reference.ParseNamed(ref)
	if err != nil {
		return nil, err
	}
	repository, err := registry.Repository(named)
	if err != nil {
		return nil, err
	}
	m, err := repository.Manifest(context.TODO(), named)
	if err != nil {
		return nil, err
	}

	for _, ref := range m.References() {
		switch ref.MediaType {
		case "application/vnd.docker.container.image.v1+json":
			continue
		case "application/vnd.docker.image.rootfs.diff.tar.gzip":
			reader, err := repository.Blob(context.TODO(), ref.Digest)
			if err != nil {
				return nil, err
			}
			err = mergeLayer(archive, reader, ignores...)
			if err != nil {
				return nil, err
			}
			continue
		default:
			return nil, fmt.Errorf("Unsupported media type in manifest: %s", ref.MediaType)
		}
	}
	return archive, nil
}

func mergeLayer(archive *fs.Archive, reader io.ReadSeekCloser, ignores ...predicates.Predicate) error {
	gzr, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gzr.Close()
	tr := tar.NewReader(gzr)
	for header, err := tr.Next(); err != io.EOF; header, err = tr.Next() {
		if err != nil {
			return err
		}
		if header.FileInfo().IsDir() {
			continue
		}
		n := fmt.Sprintf("/%s", strings.TrimPrefix(strings.TrimPrefix(header.Name, "."), "/"))
		d, err := digest.FromReader(tr)
		if err != nil {
			return err
		}
		f := fs.File{Name: n, Digest: d.String()}
		b := f.Base()
		switch {
		case b == ".wh..wh..opq":
			// Deletes the directory itself, this does not match the Image Spec behavior
			// It is okay here, because we are not interested in empty directories
			archive.Files = archive.Files.Delete(f.Path())
		case strings.HasPrefix(b, ".wh."):
			archive.Files = archive.Files.Delete(fmt.Sprintf("%s%s", f.Path(), strings.TrimPrefix(b, ".wh.")))
		default:
			ignores := predicates.Predicates(ignores)
			if !ignores.Applies(f.Name) {
				// fmt.Printf("Adding file %s to image\n", n)
				archive.Files = append(archive.Files, f)
			}
		}
	}
	return nil
}
