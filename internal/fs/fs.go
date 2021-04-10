package fs

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/opencontainers/go-digest"
)

type FileSystemIterator interface {
	Next() (*FileSystemEntry, error)
}

type FileSystemEntry struct {
	Name   string
	Reader io.Reader
}

type FileSystemInfo struct {
	Digest string
	Files  FileSlice
}

func FromIterator(f FileSystemIterator) (*FileSystemInfo, error) {
	nd := make([]string, 0)
	fs := &FileSystemInfo{
		Files: make(FileSlice, 0),
	}

	for entry, err := f.Next(); err != io.EOF; entry, err = f.Next() {
		if err != nil {
			return nil, err
		}

		dgst, err := fromReader(entry.Reader)
		if err != nil {
			return nil, err
		}

		nd = append(nd, fmt.Sprintf("%s;%s\n", entry.Name, dgst))
		fs.Files = append(fs.Files, File{Name: entry.Name, Digest: dgst})
	}

	fs.Digest = digestStrings(nd)

	return fs, nil
}

func fromReader(r io.Reader) (string, error) {
	dgst, err := digest.FromReader(r)
	if err != nil {
		return "", err
	}
	return dgst.String(), nil
}

func digestStrings(s []string) string {
	sort.Strings(s)
	sb := strings.Builder{}
	for _, entry := range s {
		sb.WriteString(entry)
	}
	return digest.FromString(sb.String()).String()
}
