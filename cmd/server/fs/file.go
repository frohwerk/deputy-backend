package fs

import (
	"fmt"
	"path/filepath"
	"strings"
)

type File struct {
	Name   string
	Digest string
}

type FileSlice []File

// Returns the last element of the path. Trailing separators are removed. If the reference is nil or the file name is empty Base returns an empty string.
func (f *File) Base() string {
	switch {
	case f == nil:
		fallthrough
	case f.Name == "":
		return ""
	default:
		return filepath.Base(f.Name)
	}
}

// Returns the Path of the file. If the reference is nil or the file name is empty Path returns an empty string.
func (f *File) Path() string {
	switch {
	case f == nil:
		return ""
	case len(f.Name) == 0:
		return ""
	default:
		i := strings.LastIndex(f.Name, "/")
		if i == -1 {
			return f.Name
		}
		return f.Name[:i+1]
	}
}

// Delete all files whose absolute path has the specified prefix.
func (files FileSlice) Delete(prefix string) FileSlice {
	for i := 0; i < len(files); i++ {
		if strings.HasPrefix(files[i].Name, prefix) {
			fmt.Printf("Deleting file %s from image\n", files[i].Name)
			if i == len(files) {
				files = files[:i]
			} else {
				files = append(files[:i], files[i+1:]...)
			}
		}
	}
	return files
}

func (files FileSlice) Len() int {
	return len(files)
}

func (files FileSlice) Less(i, j int) bool {
	return files == nil || files[i].Name < files[j].Name
}

func (files FileSlice) Swap(i, j int) {
	if files != nil {
		files[i], files[j] = files[j], files[i]
	}
}
