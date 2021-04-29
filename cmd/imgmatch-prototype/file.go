package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

type File struct {
	name   string
	digest string
}

type FileSlice []File

// Returns the Path of the file. Empty if the file pointer is nil or the name attribute is empty
func (f *File) Path() string {
	switch {
	case f == nil:
		return "nil"
	case len(f.name) == 0:
		return ""
	default:
		i := strings.LastIndex(f.name, "/")
		if i == -1 {
			return f.name
		}
		return f.name[:i+1]
	}
}

// Returns the last element of the path. Trailing separators are removed.
// If the reference is nil Base returns an empty string.
// If the file name is empty it returns "."
func (f *File) Base() string {
	switch {
	case f == nil:
		return ""
	default:
		return filepath.Base(f.name)
	}
}

// Delete all entries with the specified prefix
func (files FileSlice) Delete(prefix string) {
	for i := 0; i < len(files); i++ {
		if strings.HasPrefix(files[i].name, prefix) {
			fmt.Printf("Deleting file %s from image", files[i].name)
			files = append(files[:i], files[i+1:]...)
		}
	}
}

func (files FileSlice) Len() int {
	return len(files)
}

func (files FileSlice) Less(i, j int) bool {
	return files[i].name < files[j].name
}

func (files FileSlice) Swap(i, j int) {
	files[i], files[j] = files[j], files[i]
}
