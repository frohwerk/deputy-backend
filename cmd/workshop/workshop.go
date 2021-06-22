package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/frohwerk/deputy-backend/cmd/server/apps"
	"github.com/frohwerk/deputy-backend/cmd/workshop/dependencies"
	"github.com/frohwerk/deputy-backend/cmd/workshop/rollout"
	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/epoch"
	k8s "github.com/frohwerk/deputy-backend/internal/kubernetes"
	"github.com/frohwerk/deputy-backend/internal/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/homedir"
)

var Log logger.Logger = logger.Noop

func init() {
	logger.Default = logger.Basic(logger.LEVEL_DEBUG)
	Log = logger.Default
}

func main() {
	Log.Debug("command line: %s", os.Args[1:])
	switch len(os.Args) {
	case 1:
		Log.Fatal("missing parameter: application id")
	case 2:
		Log.Fatal("missing parameter: source unix time (or the word 'now')")
	case 3:
		Log.Fatal("missing parameter: source environment id")
	case 4:
		Log.Fatal("missing parameter: target environment id")
	}

	appId := os.Args[1]
	at := os.Args[2]
	source := os.Args[3]
	target := os.Args[4]

	before, err := parseTime(at)
	if err != nil {
		Log.Fatal("invalid parameter value 'at': %s", err)
	}

	rollout.Log = logger.Basic(logger.LEVEL_DEBUG)

	Log.Info("Source time: %v", before)

	db := database.Open()
	defer db.Close()

	appsRepo := apps.NewRepository(db)
	platforms := k8s.CreateConfigRepository(db)

	if len(appId) < 36 {
		name, row := appId, db.QueryRow(`SELECT id FROM apps WHERE name = $1`, appId)
		if err := row.Scan(&appId); err != nil {
			Log.Fatal("No application with name '%s' found", name)
		}
	}

	if len(source) < 36 {
		name, row := appId, db.QueryRow(`SELECT id FROM envs WHERE name = $1`, source)
		if err := row.Scan(&source); err != nil {
			Log.Fatal("No application with name '%s' found", name)
		}
	}

	if len(target) < 36 {
		name, row := appId, db.QueryRow(`SELECT id FROM envs WHERE name = $1`, target)
		if err := row.Scan(&target); err != nil {
			Log.Fatal("No application with name '%s' found", name)
		}
	}

	planner := rollout.Strategy(dependencies.Lookup{Store: dependencies.Cache(dependencies.DefaultDatabase(db))})

	targetEnv, err := platforms.Environment(target)
	if err != nil {
		Log.Fatal("error reading target environment configuration: %s", err)
	}

	targetApp, err := appsRepo.CurrentView(appId, target)
	if err != nil {
		Log.Fatal("error reading target application: %s", err)
	}

	sourceApp, err := appsRepo.History(appId, source, &before)
	if err != nil {
		Log.Fatal("error reading source application: %s", err)
	}

	patches, err := createPatches(sourceApp.Components, targetApp.Components)
	if err != nil {
		Log.Fatal("error creating patches for target: %s", err)
	}

	sort.Slice(patches, func(i, j int) bool { return patches[i].ComponentId > patches[j].ComponentId })
	Log.Debug("Patches before planing stage: %s", rollout.PatchList(patches))

	plan, err := planner.CreatePlan(patches)
	if err != nil {
		Log.Fatal("error creating patches for target: %s", err)
	}

	if len(plan) > -2 {
		Log.Debug("Rollout plan: %s", plan)
		return
	}

	Log.Info("Patching environment %s", target)
	for _, patch := range plan {
		target, err := targetEnv.Platform(patch.PlatformName)
		if err != nil {
			Log.Fatal("error reading target platform: %v", err)
		}

		// Maybe use context for timeout?
		complete, err := target.Apply(&patch)
		if err != nil {
			Log.Fatal("error applying patch: %v", err)
		}

		// Wait for completion
		<-complete
	}

	// TODO ./cmd/server/apps/get.go: Reuse apps view with history
	// Input: app_id, at (timestamp), from_env, to_env
	Log.Debug("TODO: add timeout for the whole thing")
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
		Log.Info("error reading cadata from %s: %s", cafile, err)
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

	Log.Debug("Components (source):")
	for _, c := range source {
		Log.Debug("%s => %s", c.Name, c.Image)
	}

	Log.Debug("Components (target):")
	for _, c := range target {
		Log.Debug("%s => %s", c.Name, c.Image)
	}

	for i := 0; i < len(source); i++ {
		source, target := source[i], target[i]
		switch {
		case source.Id != target.Id:
			return nil, fmt.Errorf("source and target must have the same components")
		case source.Platform != target.Platform:
			return nil, fmt.Errorf("source and target must use the same platform (may use different environments)")
		case source.Image == target.Image:
			Log.Debug("%v == %v", source.Image, target.Image)
			continue
		case source.Image == "":
			return nil, fmt.Errorf("source has no image specified for component %s", source.Id)
		case target.Image == "":
			return nil, fmt.Errorf("target has no image specified for component %s", source.Id)
		}
		patch := k8s.DeploymentPatch{
			ComponentId:   source.Id,
			ComponentName: source.Name,
			PlatformName:  source.Platform,
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
	case "now":
		return time.Now(), nil
	default:
		return epoch.ParseTime(t)
	}
}

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
