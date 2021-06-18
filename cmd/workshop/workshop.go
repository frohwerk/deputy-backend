package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/frohwerk/deputy-backend/cmd/server/tasks"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/homedir"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("missing parameter: image reference")
	}

	image := os.Args[1]

	cafile := "E:/projects/go/src/github.com/frohwerk/deputy-backend/certificates/minishift.crt"
	cadata, err := os.ReadFile(cafile)
	if err != nil {
		fmt.Printf("error reading cadata from %s: %s", cafile, err)
		os.Exit(1)
	}

	config := &rest.Config{
		Host:        "https://192.168.178.31:8443",
		BearerToken: "eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJteXByb2plY3QiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlY3JldC5uYW1lIjoiZGVwdXR5LXRva2VuLXJncGh6Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6ImRlcHV0eSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjgwMTFiMDk2LWFjYTktMTFlYi05YjE0LTAwMTU1ZDYzMDEwOCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpteXByb2plY3Q6ZGVwdXR5In0.VRdGoGmkesFga1GU0ooP2KbwSzuq5zb9c3mNc8j0KGYd-eFe1-39FAG4TJU2is1b0tble5SF3TB0e4x4xFlBNNEtV2jUm7htOm0le0av6KtdTaGJA3WYhLKg_BD5G8Xq9irjRZg_rp448g1Bw03yzjF-YuOeWc9T95LMcT4bGarun1QxAPAx2ZBRNZxOZe7640x1X2s3qW5XocOSRRsBmtkpC-nJ-QYvlZsRGheU8-XSGT-gy-jDKU3KFOTA4dDsZSLgkmYzK4tb1hQEYKnUbH2Jjd74dIKpgMT27a_N77TS1-b36KGltaZEBEt7kfcHXHKijXrMCzJLEHOPOCEvXw",
		TLSClientConfig: rest.TLSClientConfig{
			CAData: cadata,
		},
	}

	kube, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("error creating kubernetes client: %s", err)
	}

	tasks.Stuff(kube, image)
}

func getKubeconfig() (string, error) {
	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config"), nil
	} else {
		return "", fmt.Errorf("error loading kubeconfig: no home directory found")
	}
}
