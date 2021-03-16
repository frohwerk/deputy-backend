package main

import (
	"fmt"
	"log"
	"os"

	"github.com/frohwerk/deputy-backend/cmd/fsexample/filesystem"
	"github.com/frohwerk/deputy-backend/cmd/fsexample/registry"

	// init functions register a UnmarshalFunc with the the docker distribution client
	_ "github.com/distribution/distribution/manifest/manifestlist"
	_ "github.com/docker/distribution/manifest/ocischema"

	"github.com/opencontainers/go-digest"
)

func main() {
	if wd, err := os.Getwd(); err != nil {
		log.Fatalf("ERROR Failed to get working directory: %s\n", err)
	} else {
		log.Printf("INFO  Current working directory: %s\n", wd)
	}

	fs, err := filesystem.FromArchive("temp/app.tar.gz")
	if err != nil {
		log.Fatalf("ERROR Failed to read archive: %s\n", err)
	}

	fmt.Printf("Archive file system\n")
	fmt.Printf("===================\n")
	for name, dig := range fs {
		fmt.Printf("%s  %s\n", dig, name)
	}

	reg, err := registry.New("http://ocrproxy-myproject.192.168.178.31.nip.io/v2")
	if err != nil {
		log.Fatalf("ERROR Failed to create registry client: %s\n", err)
	}

	r, err := reg.Repo("myproject/node-hello-world")
	if err != nil {
		log.Fatalf("ERROR Failed to create repository client: %s\n", err)
	}

	ifs, err := filesystem.FromImage(r, digest.Digest("sha256:3bf137c335a2f7f9040eef6c2093abaa273135af0725fdeea5c4009a695d840f"))
	if err != nil {
		log.Fatalf("ERROR Error reading image file system: %s\n", err)
	}

	fmt.Print("\n")
	fmt.Print("Image file system\n")
	fmt.Print("=================\n")
	for name, dig := range ifs {
		fmt.Printf("%s  %s\n", dig, name)
	}

	if !ifs.Contains(fs) {
		fmt.Printf("INFO  Image file system does not contain archive file system!")
	}
}
