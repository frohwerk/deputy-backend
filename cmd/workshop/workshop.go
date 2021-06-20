package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/frohwerk/deputy-backend/cmd/server/apps"
	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/epoch"
	k8s "github.com/frohwerk/deputy-backend/internal/kubernetes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/homedir"
)

type byId []apps.Component

func (s byId) Len() int {
	return len(s)
}

func (s byId) Less(i, j int) bool {
	return s[i].Id < s[j].Id
}

func (s byId) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type PatchSpec struct {
	Template PatchTemplate `json:"template,omitempty"`
}

type PatchTemplate struct {
	Spec corev1.PodSpec `json:"spec,omitempty"`
}

type DeploymentPatch struct {
	Component string    `json:"-"`
	Platform  string    `json:"-"`
	Spec      PatchSpec `json:"spec,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("missing parameter: image reference")
	}

	appId := os.Args[1]
	at := os.Args[2]
	source := os.Args[3]
	target := os.Args[4]

	before, err := parseTime(at)
	if err != nil {
		log.Fatalf("invalid parameter value 'at': %s", err)
	}

	fmt.Printf("Source time: %v\n", before)

	db := database.Open()
	defer db.Close()

	repo := apps.NewRepository(db)
	platforms := k8s.CreateConfigRepository(db)

	targetEnv, err := platforms.Environment(target)
	if err != nil {
		log.Fatalf("error reading target environment configuration: %s", err)
	}

	targetApp, err := repo.CurrentView(appId, target)
	if err != nil {
		log.Fatalf("error reading target application: %s", err)
	}

	sourceApp, err := repo.History(appId, source, &before)
	if err != nil {
		log.Fatalf("error reading source application: %s", err)
	}

	patches, err := createPatches(sourceApp.Components, targetApp.Components)
	if err != nil {
		log.Fatalf("error creating patches for target: %s", err)
	}

	fmt.Printf("Patching environment %s\n", target)
	for _, patch := range patches {
		target, err := targetEnv.Platform(patch.Platform)
		if err != nil {
			log.Fatalf("error reading target platform: %v", err)
		}

		// Maybe use context for timeout?
		complete, err := target.Apply(&patch)
		if err != nil {
			log.Fatalf("error applying patch: %v", err)
		}

		// Wait for completion
		<-complete
	}

	// TODO ./cmd/server/apps/get.go: Reuse apps view with history
	// Input: app_id, at (timestamp), from_env, to_env
	fmt.Println("TODO: add timeout for the whole thing")
	// For each component (determine a reasonable order!):
	//
}

func getKubeconfig() (string, error) {
	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config"), nil
	} else {
		return "", fmt.Errorf("error loading kubeconfig: no home directory found")
	}
}

func createKubernetesClient() (*kubernetes.Clientset, error) {
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

	return kubernetes.NewForConfig(config)
}

func createPatches(source, target []apps.Component) ([]k8s.DeploymentPatch, error) {
	if len(source) != len(target) {
		return nil, fmt.Errorf("source and target must have the same components")
	}

	patches := []k8s.DeploymentPatch{}
	sort.Sort(byId(source))
	sort.Sort(byId(target))

	fmt.Println("Components (source):")
	for _, c := range source {
		fmt.Printf("%s => %s\n", c.Name, c.Image)
	}

	fmt.Println("Components (target):")
	for _, c := range target {
		fmt.Printf("%s => %s\n", c.Name, c.Image)
	}

componentLoop:
	for i := 0; i < len(source); i++ {
		source, target := source[i], target[i]
		switch {
		case source.Id != target.Id:
			return nil, fmt.Errorf("source and target must have the same components")
		case source.Platform != target.Platform:
			return nil, fmt.Errorf("source and target must use the same platform (may use different environments)")
		case source.Image == target.Image:
			fmt.Printf("%v == %v\n", source.Image, target.Image)
			continue componentLoop
		case source.Image == "":
			return nil, fmt.Errorf("source has no image specified for component %s", source.Id)
		case target.Image == "":
			return nil, fmt.Errorf("target has no image specified for component %s", source.Id)
		}
		patch := k8s.DeploymentPatch{
			Component: source.Name,
			Platform:  source.Platform,
			Spec: k8s.DeploymentPatchSpec{
				Template: k8s.PodTemplatePatch{
					Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: source.Name, Image: source.Image}}},
				},
			},
		}
		patches = append(patches, patch)
	}

	return patches, nil
}

func parseTime(t string) (time.Time, error) {
	switch t {
	case "*":
		return time.Now(), nil
	default:
		return epoch.ParseTime(t)
	}
}
