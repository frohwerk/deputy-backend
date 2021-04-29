package main

import (
	"fmt"
	"log"
	"sort"

	"github.com/frohwerk/deputy-backend/cmd/imgmatch/predicates"
	"github.com/frohwerk/deputy-backend/cmd/server/images"
	"github.com/frohwerk/deputy-backend/internal/fs/img"
)

func main() {
	registry := &images.RemoteRegistry{BaseUrl: "http://ocrproxy-myproject.192.168.178.31.nip.io"}
	fs, err := img.FromImage("172.30.1.1:5000/myproject/node-hello-world:1.0.3", registry, predicates.Prefix("/nodejs/", "/usr/lib/x86_64-linux-gnu/gconv/", "/usr/share/doc/", "/usr/share/man/", "/usr/share/zoneinfo/"))
	if err != nil {
		log.Println(err)
	} else {
		sort.Sort(fs.Files)
		for _, f := range fs.Files {
			fmt.Println(f.Name, f.Digest)
		}
	}
}
