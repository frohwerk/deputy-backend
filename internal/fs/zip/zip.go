package zip

import (
	"archive/zip"
	"io"

	"github.com/frohwerk/deputy-backend/internal/fs"
)

func FromZipReader(name string, zr *zip.Reader) (*fs.Archive, error) {
	fsi, err := fs.FromIterator(&zipFsIterator{zr: zr, i: 0, len: len(zr.File)})
	if err != nil {
		return nil, err
	}
	return &fs.Archive{Name: name, FileSystemInfo: fsi}, nil
}

type zipFsIterator struct {
	i    int
	len  int
	zr   *zip.Reader
	prev io.Closer
}

func (zfs *zipFsIterator) Next() (*fs.FileSystemEntry, error) {
	if zfs.prev != nil {
		zfs.prev.Close()
	}
	for zfs.i < zfs.len {
		file := zfs.zr.File[zfs.i]
		if file.FileInfo().IsDir() {
			zfs.i++
			continue
		}
		r, err := file.Open()
		if err != nil {
			return nil, err
		}
		zfs.prev = r
		zfs.i++

		return &fs.FileSystemEntry{
			Name:   file.Name,
			Reader: r,
		}, nil
	}
	return nil, io.EOF
}
