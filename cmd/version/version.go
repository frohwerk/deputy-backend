package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	root := &cobra.Command{RunE: run}
	if err := root.Execute(); err != nil {
		fmt.Println(err)
	}
}

func run(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Missing argument: version")
	}
	version := args[0]

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: filepath.Join(home, ".kube", "config")}
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	clientConfig, err := config.ClientConfig()
	if err != nil {
		return err
	}

	client, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return err
	}

	namespace, _, err := config.Namespace()
	if err != nil {
		return err
	}
	fmt.Println("Using namespace", namespace)

	spec := DeploymentPatch{
		Spec: DeploymentSpecPatch{
			Template: PodTemplatePatch{
				Spec: PodSpecPatch{
					Containers: []ContainerPatch{
						{Name: "node-hello-world", Image: fmt.Sprintf("172.30.1.1:5000/myproject/node-hello-world:%s", version)},
					},
				},
			},
		},
	}

	patch, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	fmt.Print("Applying patch:\n", string(patch), "\n")

	result, err := client.AppsV1().Deployments(namespace).Patch("node-hello-world", types.StrategicMergePatchType, patch)
	if err != nil {
		return err
	}

	for _, c := range result.Spec.Template.Spec.Containers {
		if c.Name == "node-hello-world" {
			fmt.Println("Patched container - current image:", c.Image)
		}
	}

	return nil
}
