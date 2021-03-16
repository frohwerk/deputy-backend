package fs

import (
	"io"
)

type FileSystem interface {
	Next() (*FileSystemEntry, error)
}

type FileSystemEntry struct {
	Name   string
	Reader io.Reader
}
