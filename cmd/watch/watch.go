package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"

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

	podsWatch, err := client.CoreV1().Pods("myproject").Watch(meta.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %s", err)
	}
	pods := podsWatch.ResultChan()

	deploymentsWatch, err := client.AppsV1().Deployments("myproject").Watch(meta.ListOptions{FieldSelector: "metadata.name=node-hello-world"})
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %s", err)
	}
	deployments := deploymentsWatch.ResultChan()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

watchloop:
	for i := 1; ; {
		select {
		case evt := <-pods:
			switch o := evt.Object.(type) {
			case *core.Pod:
				for _, owner := range o.ObjectMeta.OwnerReferences {
					fmt.Println("Pod owner:", owner.Name, "(", owner.Kind, ")")
					if owner.Kind != "ReplicaSet" {
						continue
					}
					rs, err := client.AppsV1().ReplicaSets("myproject").Get(owner.Name, meta.GetOptions{})
					if err != nil {
						fmt.Fprintf(os.Stderr, "error getting replica set for pod: %s\n", err)
						continue
					}
					for _, rsown := range rs.ObjectMeta.OwnerReferences {
						fmt.Println("ReplicaSet owner:", rsown.Name, "(", rsown.Kind, ")")
					}
				}
				if !strings.HasPrefix(o.ObjectMeta.Name, "node-hello-world") {
					continue watchloop
				}
				name := fmt.Sprintf("%s/%05d-%s-%s.yaml", directory, i, evt.Type, o.ObjectMeta.Name)
				if err := writeFile(name, o); err != nil {
					return err
				}
				i++
			default:
				fmt.Fprintf(os.Stderr, "Unexpected event object: %T", o)
			}
			if err != nil {
				return fmt.Errorf("error encoding to yaml: %s\n", err)
			}
		case evt := <-deployments:
			switch o := evt.Object.(type) {
			case *apps.Deployment:
				if !strings.HasPrefix(o.ObjectMeta.Name, "node-hello-world") {
					fmt.Println("Unexpected object name:", o.ObjectMeta.Name)
					continue watchloop
				}

				if evt.Type == watch.Modified {
					fmt.Println("Label selectors:")
					selector := o.Spec.Selector.MatchLabels
					for k, v := range selector {
						fmt.Println(k, ":", v)
					}
					podsWatch, err := client.CoreV1().Pods("myproject").Watch(meta.ListOptions{LabelSelector: labels.Set(selector).String()})
					if err != nil {
						fmt.Fprintf(os.Stderr, "error creating Watch for pods: %v\n", err)
						continue watchloop
					}

					replicas := *o.Spec.Replicas
					for replicas > 0 {
						evt := <-podsWatch.ResultChan()
						pod, ok := evt.Object.(*core.Pod)
						if !ok {
							fmt.Fprintf(os.Stderr, "unexpected object type on pod watch: %T\n", evt.Object)
							continue
						}
						image := pod.Spec.Containers[0].Image
						fmt.Printf("Pod %s with image %s\n", evt.Type, image)
						if evt.Type == watch.Deleted {
							replicas--
						}
					}
				}

				name := fmt.Sprintf("%s/%05d-%s-%s.yaml", directory, i, evt.Type, o.ObjectMeta.Name)
				fmt.Println("Writing event to", name)
				if err := writeFile(name, o); err != nil {
					return err
				}
				i++
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
				deploymentsWatch.Stop()
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
