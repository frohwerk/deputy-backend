package zip

import (
	"archive/zip"
	"io"

	"github.com/frohwerk/deputy-backend/cmd/server/fs"
)

func FromZipReader(name string, zr *zip.Reader) (*fs.FileSystemInfo, error) {
	return fs.FromFilesystem(name, &zipFileSystem{zr: zr, i: 0, len: len(zr.File)})
}

type zipFileSystem struct {
	zr   *zip.Reader
	i    int
	len  int
	prev io.Closer
}

func (zfs *zipFileSystem) Next() (*fs.FileSystemEntry, error) {
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
