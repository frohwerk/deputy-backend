package fs

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/opencontainers/go-digest"
)

type FileSystemDigests struct {
	Name        string
	Digest      string
	FileNames   map[string][]string
	FileDigests map[string]string
}

func FromFilesystem(name string, f FileSystem) (*FileSystemDigests, error) {
	nd := make([]string, 0)
	fs := &FileSystemDigests{
		Name:        name,
		FileDigests: make(map[string]string),
		FileNames:   make(map[string][]string),
	}

	for entry, err := f.Next(); err != io.EOF; entry, err = f.Next() {
		if err != nil {
			return nil, err
		}

		dgst, err := fromReader(entry.Reader)
		if err != nil {
			return nil, err
		}

		fs.FileDigests[entry.Name] = dgst
		if names, ok := fs.FileNames[dgst]; ok {
			fs.FileNames[dgst] = append(names, entry.Name)
		} else {
			fs.FileNames[dgst] = []string{entry.Name}
		}
		nd = append(nd, fmt.Sprintf("%s;%s\n", entry.Name, dgst))
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
