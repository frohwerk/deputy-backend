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

// A File matches another File when digests and file names (ignoring the path) are equal
func (f *File) Matches(ref *File) bool {
	return f.digest == ref.digest && filepath.Base(f.name) == filepath.Base(ref.name)
}

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

func (f *File) Base() string {
	switch {
	case f == nil:
		return ""
	default:
		return filepath.Base(f.name)
	}
}

func (files FileSlice) Search(ref *File) (int, bool) {
	if ref == nil {
		return -1, false
	}
	for i, f := range files {
		if f.Matches(ref) {
			return i, true
		}
	}
	return -1, false
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
