package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/spf13/cobra"

	apps "k8s.io/api/apps/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	yaml "gopkg.in/yaml.v2"
)

var (
	namespace string
	directory string

	command = &cobra.Command{RunE: run}
)

func init() {
	command.Flags().StringVarP(&directory, "outputDirectory", "o", "", "directory for output files")
}

func main() {
	command.Use = "watch <namespace>"
	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	switch {
	case len(args) != 1:
		fmt.Println("Please specify the namespace to watch")
		return cmd.Usage()
	case directory == "":
		fmt.Println("Please specify the output directory")
		return cmd.Usage()
	default:
		namespace = args[0]
	}
	kubeconfig, err := getKubeconfig()
	if err != nil {
		return fmt.Errorf("failed to load $HOME/.kube/config: %s", err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to parse $HOME/.kube/config: %s", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %s", err)
	}

	watch, err := client.AppsV1().Deployments("myproject").Watch(meta.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %s", err)
	}
	deployments := watch.ResultChan()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	for i := 1; ; i++ {
		select {
		case evt := <-deployments:
			switch o := evt.Object.(type) {
			case *apps.Deployment:
				name := fmt.Sprintf("%s/%05d-%s.yaml", directory, i, o.ObjectMeta.Name)
				fmt.Println("Writing event to", name)
				if err := writeFile(name, o); err != nil {
					return err
				}
			default:
				fmt.Fprintf(os.Stderr, "Unexpected event object: %T", o)
			}
			if err != nil {
				return fmt.Errorf("error encoding to yaml: %s\n", err)
			}
		case sig := <-signals:
			switch sig {
			case os.Interrupt:
				fallthrough
			case os.Kill:
				fmt.Println("And now my watch has ended")
				watch.Stop()
				return nil
			default:
				fmt.Fprintf(os.Stderr, "unexpected signal: %s\n", sig)
			}
		}
	}
}

func getKubeconfig() (string, error) {
	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config"), nil
	} else {
		return "", fmt.Errorf("error loading kubeconfig: no home directory found")
	}
}

func writeFile(fname string, v interface{}) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewEncoder(f).Encode(v)
}
