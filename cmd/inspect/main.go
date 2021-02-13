package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/frohwerk/deputy-backend/internal/kubernetes"
)

func main() {
	cluster, err := kubernetes.WithDefaultConfig()
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	artifacts, err := cluster.GetComponents()
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	for _, artifact := range artifacts {
		if artifact.Image == "172.30.1.1:5000/myproject/ocrproxy" {
			log.Printf("TRACE fetching manifest for image '%v'\n", artifact.Image)

			m, err := GetManifest(artifact.Image)
			if err != nil {
				log.Printf("error fetching manifest: %v\n", err)
				continue
			}

			dig, err := m.GetArtifact("ocrproxy")
			if err != nil {
				log.Printf("error fetching artifact metadata: %v\n", err)
				continue
			}

			log.Printf("INFO  digest for artifact 'ocrproxy': %v\n", dig)
		}
		log.Printf("DEBUG %v\n", artifact.Image)
	}
}

func (m *Manifest) GetArtifact(s string) (string, error) {
	if len(m.Layers) == 0 {
		return "", fmt.Errorf("image has no file system layers")
	}

	l := m.Layers[len(m.Layers)-1]
	uri := fmt.Sprintf("http://ocrproxy-myproject.192.168.178.31.nip.io/v2/%v/%v/blobs/%v", m.Repository, m.Name, l.Digest)
	req, err := http.NewRequest("GET", uri, nil)
	req.Header.Set("Accept", "application/vnd.docker.image.rootfs.diff.tar.gzip")
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	} else {
		defer resp.Body.Close()
	}

	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", err
	}

	tr := tar.NewReader(gzr)

	digest := "unknown"
	for h, err := tr.Next(); err != io.EOF; h, err = tr.Next() {
		if err != nil {
			return "", err
		}
		log.Printf("TRACE found entry in file system layer: '%v'\n", h.Name)
		buf, err := ioutil.ReadAll(tr)
		if err != nil {
			return "", err
		}
		if h.Name == s {
			buf := sha256.Sum256(buf)
			digest = hex.EncodeToString(buf[:])
		} else if strings.Contains(h.Name, s) {
			log.Printf("WARN  name matches but is not equal: h.Name = '%v' s = '%v'\n", h.Name, s)
			buf := sha256.Sum256(buf)
			digest = hex.EncodeToString(buf[:])
		}
	}

	// dir, err := ioutil.TempDir("", "deputy-*")
	// if err != nil {
	// 	return "", err
	// } else {
	// 	defer os.RemoveAll(dir)
	// }
	return digest, nil
}

func GetManifest(s string) (*Manifest, error) {
	ref, err := ParseRef(s)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("http://ocrproxy-myproject.192.168.178.31.nip.io/v2/%v/%v/manifests/%v", ref.Repository, ref.Name, ref.Tag)
	req, err := http.NewRequest("GET", uri, nil)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	} else {
		defer resp.Body.Close()
	}

	manifest := &Manifest{ImageRef: ref}
	dec := json.NewDecoder(resp.Body)
	if err = dec.Decode(manifest); err != nil {
		return nil, err
	}

	return manifest, nil
}

func ParseRef(s string) (*ImageRef, error) {
	var imageRef *ImageRef

	parts := strings.Split(s, "/")
	switch len(parts) {
	case 3:
		imageRef = &ImageRef{Host: parts[0], Repository: parts[1], Name: parts[2]}
	case 2:
		imageRef = &ImageRef{Host: "docker.io", Repository: parts[0], Name: parts[1]}
	case 1:
		imageRef = &ImageRef{Host: "docker.io", Repository: "library", Name: parts[0]}
	default:
		return nil, fmt.Errorf("invalid image ref: '%v'", s)
	}

	if strings.Contains(imageRef.Name, ":") {
		parts := strings.SplitN(imageRef.Name, ":", 2)
		imageRef.Name = parts[0]
		imageRef.Tag = parts[1]
	} else {
		imageRef.Tag = "latest"
	}

	return imageRef, nil
}

type ImageRef struct {
	Host       string
	Repository string
	Name       string
	Tag        string
}

type Reference struct {
	MediaType string `json:"mediaType"`
	Size      uint   `json:"size"`
	Digest    string `json:"digest"`
}

type Manifest struct {
	*ImageRef
	SchemaVersion uint        `json:"schemaVersion"`
	MediaType     string      `json:"mediaType"`
	Config        Reference   `json:"config"`
	Layers        []Reference `json:"layers"`
}
