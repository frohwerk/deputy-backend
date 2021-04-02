package img

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"sort"

	tarfs "github.com/frohwerk/deputy-backend/cmd/server/fs/tar"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/frohwerk/deputy-backend/cmd/server/fs"
	"github.com/opencontainers/go-digest"
)

var _ fs.FileSystem = &image{}

type ImageRepository struct {
	distribution.ManifestService
	distribution.BlobProvider
}

func (r *ImageRepository) FromImage(ctx context.Context, d digest.Digest) (*fs.FileSystemInfo, error) {
	m, err := r.ManifestService.Get(ctx, d)
	if err != nil {
		return nil, err
	}
	switch m := m.(type) {
	case *schema2.DeserializedManifest:
		return r.fromManifestV2(ctx, m)
	default:
		return nil, fmt.Errorf("Unsupported manifest type: %T", m)
	}
}

// TODO: Modify function to create a FileSystemInfo (maybe? or more like imgmatch?)
func (repo *ImageRepository) fromManifestV2(ctx context.Context, m *schema2.DeserializedManifest) (*fs.FileSystemInfo, error) {
	img := &image{}
	// Possible optimization: iterate in reverse order, remember whiteouts and skip matching entries in lower layers
	for _, layer := range m.Layers {
		r, err := repo.BlobProvider.Open(ctx, layer.Digest)
		if err != nil {
			return nil, err
		}
		defer r.Close()

		gzr, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		defer gzr.Close()

		tfs, err := tarfs.FromTarReader(layer.Digest.String(), tar.NewReader(gzr))
		if err != nil {
			return nil, err
		}

		for name, digest := range tfs.FileDigests {
			img.files = append(img.files, file{name, digest})
		}
	}
	// Sort to simplify matching later
	sort.Slice(img.files, func(i, j int) bool { return img.files[i].name < img.files[j].name })
	return nil, nil
}

type file struct {
	name   string
	digest string
}

type image struct {
	files []file
}

func (img *image) Next() (*fs.FileSystemEntry, error) {
	return nil, io.EOF
}

// type image struct {
// 	ImageRepository
// 	layers []distribution.Descriptor
// 	reader distribution.ReadSeekCloser
// 	nextLayer int
// }

// func (img *image) Next() (*fs.FileSystemEntry, error) {
// 	var err error
// 	if img.reader == nil {
// 		img.reader, err = img.ImageRepository.BlobProvider.Open(context.TODO(), img.layers[img.nextLayer].Digest)
// 		if err != nil {
// 			return nil, err
// 		}

// 		img.nextLayer++
// 	}
// 	// if img.nextLayer > len(img.layers)-1 {
// 	// 	return nil, io.EOF
// 	// }
// 	return nil, io.EOF
// }
