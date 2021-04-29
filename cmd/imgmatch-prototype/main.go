package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/frohwerk/deputy-backend/cmd/fsexample/registry"
	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/util"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
)

var (
	rootcmd = &cobra.Command{RunE: run}
)

func main() {
	rootcmd.Use = os.Args[0]
	if err := rootcmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func run(cmd *cobra.Command, args []string) error {
	// TODO: Debug and check if it's working!!!
	db, err := sql.Open("postgres", "postgres://deputy:!m5i4e3h2e1g@localhost:5432/deputy?sslmode=disable")
	if err != nil {
		return err
	}
	defer db.Close()
	fstore := database.NewFileStore(db)
	astore := database.NewFileStore(db)

	reg, err := registry.New("http://ocrproxy-myproject.192.168.178.31.nip.io/v2")
	if err != nil {
		return err
	}

	// repo, err := reg.Repo("myproject/node-hello-world")
	// repo, err := reg.Repo("myproject/jetty-hello-world")
	repo, err := reg.Repo("myproject/ocrproxy")
	if err != nil {
		return err
	}

	ctx := context.Background()
	ms, err := repo.Manifests(ctx)
	if err != nil {
		return err
	}
	// node-hello-world
	// m, err := ms.Get(ctx, digest.Digest("sha256:3bf137c335a2f7f9040eef6c2093abaa273135af0725fdeea5c4009a695d840f"))
	// jetty-hello-world
	// m, err := ms.Get(ctx, digest.Digest("sha256:f1966dbfe1d5af2f0fe5779025368aa42883ba7a188a590f64b964e0fd01eeb3"))
	// ocrproxy
	m, err := ms.Get(ctx, digest.Digest("sha256:70272a19cf1abfc1ca938a467efdc517151d50ed08589997e6faf1bd9e24c1ca"))
	if err != nil {
		return err
	}
	switch m := m.(type) {
	case *schema2.DeserializedManifest:
		fs, err := processManifestV2(ctx, repo, m)
		if err != nil {
			return err
		}
		sort.Sort(fs)
		a, err := FindArtifacts(fs, astore, fstore)
		if err != nil {
			return err
		}
		for _, f := range a {
			fmt.Printf("%s %s %s\n", f.Id, f.Name, f.Digest)
		}
	default:
		return fmt.Errorf("Unsupported manifest type: %T", m)
	}

	return nil
}

func processManifestV2(ctx context.Context, repo distribution.Repository, manifest *schema2.DeserializedManifest) (FileSlice, error) {
	fs := FileSlice{}
	// TODO: Verify optimization: Skip first layer, since it contains only the root filesystem (many unnecessary files)
	for _, layer := range manifest.Layers[1:] {
		// TODO Remove debug shortcut:
		// if layer.Digest.String() != "sha256:0ca190b54e2e5c4cb72344a2d135a0b0db388e6ce19bcbea7af55955946d5b9c" {
		// 	continue
		// }
		r, err := repo.Blobs(ctx).Open(ctx, layer.Digest)
		if err != nil {
			return nil, err
		}
		defer r.Close()

		gzr, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		defer gzr.Close()

		tr := tar.NewReader(gzr)
		for header, err := tr.Next(); err != io.EOF; header, err = tr.Next() {
			if err != nil {
				return nil, err
			}
			if header.FileInfo().IsDir() {
				continue
			}
			n := fmt.Sprintf("/%s", strings.TrimPrefix(strings.TrimPrefix(header.Name, "."), "/"))
			d, err := digest.FromReader(tr)
			if err != nil {
				return nil, err
			}
			f := File{name: n, digest: d.String()}
			b := f.Base()
			switch {
			case b == ".wh..wh..opq":
				// Deletes the directory itself, this does not match the Image Spec behavior
				// It is okay here, because we are not interested in empty directories
				fs.Delete(f.Path())
			case strings.HasPrefix(b, ".wh."):
				fs.Delete(fmt.Sprintf("%s/%s", f.Path(), strings.TrimPrefix(b, ".wh.")))
			default:
				fs = append(fs, f)
			}
		}
	}
	return fs, nil
}

func FindArtifacts(fs []File, fileStore database.FileDigestFinder, archiveStore database.ArchiveLookup) ([]database.File, error) {
	result := []database.File{}
	for _, f := range fs {
		files, err := fileStore.FindByDigest(f.digest)
		switch {
		case err != nil:
			return nil, err
		case len(files) == 0:
			continue
		}
		result = append(result, files...)
	}
	if len(result) > 0 {
		return result, nil
	}
	strength := 0
	scanned := make(set)
	cb := &util.ControlBreak{}
	for i, f := range fs {
		if cb.IsBreak(f.Path()) {
			scanned.clear()
		}
		// Match using file digests
		archives, err := archiveStore.FindByContent(&database.File{Name: f.name, Digest: f.digest})
		if err != nil {
			return nil, err
		}
		matches := 0
		// Match using archive file contents
	archiveloop:
		for _, archive := range archives {
			if scanned.contains(archive.Id) {
				continue archiveloop
			}
			sort.Sort(archive.Files)
			if archive.Files[0].Digest != f.digest {
				return nil, fmt.Errorf("the first item in the archive should match the first matching file in the directory")
			}
			for j, k := 0, i+matches; j < len(archive.Files) && k < len(fs); {
				// TODO: path should be set for the first match only! everything else should be relative to this path
				path := strings.TrimSuffix(fs[k].name, archive.Files[j].Name)
				name := fmt.Sprintf("%s%s", path, archive.Files[j].Name)
				fmt.Printf("Matching file %s with %s\n", name, fs[k].name)
				switch {
				case name > fs[k].name:
					k++
				case name < fs[k].name:
					fallthrough
				case archive.Files[j].Digest != fs[k].digest:
					fmt.Printf(">>>> Unmatched file: %s from archive '%s' not found in image\n", name, archive.File.Name)
					scanned.put(archive.Id)
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
				scanned.put(archive.Id)
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

// func match_incomplete_impl() {
// 		sort.Sort(fs)

// 		fmt.Println()
// 		fmt.Printf("Layer %s", layer.Digest)
// 		fmt.Println()
// 		fmt.Println(strings.Repeat("=", 77))
// 		for offset, name := range names {
// 			d := digests[name]
// 			// fmt.Printf("%s %s\n", digests[name], name)
// 			files, err := fstore.FindByDigest(d)
// 			if err != nil {
// 				fmt.Fprintf(os.Stderr, "error reading files table for file %s (sha256: %s)\n", name, d)
// 				continue
// 			}
// 			if len(files) == 0 {
// 				// fmt.Printf("File %s (sha256: %s) not found in database\n", name, d)
// 				continue
// 			}
// 			for _, file := range files {
// 				fmt.Printf("file match: %s = %s (%s)\n", name, file.Id, file.Name)
// 				if file.Parent != "" {
// 					p, err := fstore.Get(file.Parent)
// 					if err != nil {
// 						fmt.Fprintf(os.Stderr, "error reading file table with file_id %s\n", file.Parent)
// 						continue
// 					}
// 					path := filepath.Dir(name)
// 					fmt.Printf("archive %s might be located at %s\n", p.Name, path)
// 					files, err := fstore.FindByParent(file.Parent)
// 					if err != nil {
// 						fmt.Fprintf(os.Stderr, "error reading file table with file_parent %s\n", file.Parent)
// 						continue
// 					}
// 					sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
// 					img := image{names: names[offset:], digests: digests}
// 					if img.containsAll(p.Name, len(files), func(i int) string { return files[i].Name }, func(i int) string { return files[i].Digest }) {
// 						fmt.Printf("Content of archive %s found at %s\n", p.Name, path)
// 					}
// 					// Since the names list is sorted it is not necessary to to a full scan, directories and contained files will appear in natural order
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

type image struct {
	names   []string
	digests map[string]string
}

type AccessorFunc func(int) string

func (img *image) containsAll(archive string, size int, name, digest AccessorFunc) bool {
	for i := 0; i < size; i++ {
		archname := name(i)
		imgname := img.names[i]
		if !strings.HasSuffix(imgname, archname) {
			fmt.Printf("%s:%s != image:%s - path mismatch\n", archive, archname, imgname)
			return false
		}
		archdigest := digest(i)
		imgdigest := img.digests[imgname]

		switch {
		case !strings.HasSuffix(imgname, archname):
			fmt.Printf("%s:%s (%s) != image:%s (%s) - path mismatch\n", archive, archname, archdigest, imgname, imgdigest)
			return false
		case archdigest != imgdigest:
			fmt.Printf("%s:%s (%s) != image:%s (%s) - digest mismatch\n", archive, archname, archdigest, imgname, imgdigest)
			return false
		}
	}
	return true
}
