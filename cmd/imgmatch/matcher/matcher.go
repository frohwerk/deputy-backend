package matcher

import (
	"fmt"
	"sort"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/frohwerk/deputy-backend/cmd/imgmatch/predicates"
	"github.com/frohwerk/deputy-backend/cmd/server/images"
	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/fs"
	imgfs "github.com/frohwerk/deputy-backend/internal/fs/img"
	"github.com/frohwerk/deputy-backend/internal/util"
)

var fsignore = predicates.Prefix("/nodejs/", "/usr/lib/x86_64-linux-gnu/gconv/", "/usr/share/doc/", "/usr/share/man/", "/usr/share/zoneinfo/")

func init() {
	fmt.Println("cmd/imgmatch/matcher/matcher.go - TODO: Replace static fsignore with customizable one")
}

type Matcher interface {
	Match(string) ([]database.File, error)
}

type matcher struct {
	archiveStore database.ArchiveLookup
	fileStore    database.FileDigestFinder
	registry     images.Registry
}

func (m *matcher) Match(ref string) ([]database.File, error) {
	fs, err := imgfs.FromImage(ref, m.registry, fsignore)
	if err != nil {
		return nil, err
	}

	matches, err := m.find(fs.Files)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func New(archiveStore database.ArchiveLookup, fileStore database.FileDigestFinder, registry images.Registry) Matcher {
	return &matcher{archiveStore, fileStore, registry}
}

func (m *matcher) find(fs []fs.File) ([]database.File, error) {
	result, err := m.findArchive(fs)
	if err != nil {
		return nil, err
	}
	if len(result) > 0 {
		return result, nil
	}
	return m.findByContent(fs)
}

func (m *matcher) findArchive(fs []fs.File) ([]database.File, error) {
	result := []database.File{}
	for _, f := range fs {
		files, err := m.fileStore.FindByDigest(f.Digest)
		switch {
		case err != nil:
			return nil, err
		case len(files) == 0:
			continue
		}
		result = append(result, files...)
	}
	return result, nil
}

func (m *matcher) findByContent(fs fs.FileSlice) ([]database.File, error) {
	sort.Sort(fs)
	result := []database.File{}
	strength := 0
	scanned := make(util.Set)
	cb := &util.ControlBreak{}
	for i, f := range fs {
		if cb.IsBreak(f.Path()) {
			scanned.Clear()
		}
		// Match using file name and digest
		archives, err := m.archiveStore.FindByContent(&database.File{Name: f.Name, Digest: f.Digest})
		if err != nil {
			return nil, err
		}
		matches := 0
		// Match using archive file contents
	archiveloop:
		for _, archive := range archives {
			if scanned.Contains(archive.Id) {
				continue archiveloop
			}
			sort.Sort(archive.Files)
			if archive.Files[0].Digest != f.Digest {
				return nil, fmt.Errorf("the first item in the archive should match the first matching file in the directory")
			}
			for j, k := 0, i+matches; j < len(archive.Files) && k < len(fs); {
				// TODO: path should be set for the first match only! everything else should be relative to this path
				path := strings.TrimSuffix(fs[k].Name, archive.Files[j].Name)
				name := fmt.Sprintf("%s%s", path, archive.Files[j].Name)
				fmt.Printf("Matching file %s with %s\n", name, fs[k].Name)
				switch {
				case name > fs[k].Name:
					k++
				case name < fs[k].Name:
					fallthrough
				case archive.Files[j].Digest != fs[k].Digest:
					fmt.Printf(">>>> Unmatched file: %s from archive '%s' not found in image\n", name, archive.File.Name)
					scanned.Put(archive.Id)
					continue archiveloop
				default:
					fmt.Printf("Found file %s at path %s\n", archive.Files[j].Name, path)
					matches++
					k++
					j++
				}
			}
			if matches == len(archive.Files) {
				fmt.Printf("Found all files, the archive '%s' is a match\n", archive.Id)
				scanned.Put(archive.Id)
				if matches > strength {
					strength, matches = matches, 0
					result = []database.File{archive.File}
				} else {
					fmt.Printf("Ignoring archive '%s', since '%s' is a better match\n", archive.Id, result[0].Id)
				}
			} else {
				fmt.Printf(">>>> Archive '%s' has %v unmatched files\n", archive.Id, len(archive.Files)-matches)
			}
		}
	}
	return result, nil
}

func parse(s string) (reference.Named, error) {
	ref, err := reference.ParseAnyReference(s)
	if err != nil {
		return nil, err
	}
	switch ref := ref.(type) {
	case reference.Named:
		return ref, nil
	default:
		return nil, fmt.Errorf("Unsupported reference type: %T", ref)
	}
}
