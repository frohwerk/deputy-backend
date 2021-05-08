package kubernetes

import (
	"fmt"
	"path/filepath"

	"github.com/frohwerk/deputy-backend/internal"
	"github.com/frohwerk/deputy-backend/pkg/api"

	appsv1 "k8s.io/api/apps/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type cluster struct {
	client *k8s.Clientset
}

func WithDefaultConfig() (*cluster, error) {
	// TODO: How to detect in-cluster configuration and out-of-cluster configuration?
	// config, err := rest.InClusterConfig()
	kubeconfig, err := getKubeconfig()
	if err != nil {
		return nil, err
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubernetes configuration: %s\n", err)
	}

	client, err := k8s.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubernetes configuration: %s\n", err)
	}

	return &cluster{client}, nil
}

func getKubeconfig() (string, error) {
	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config"), nil
	} else {
		return "", fmt.Errorf("error loading kubeconfig: no home directory found")
	}
}

func (c *cluster) GetComponents() ([]api.Component, error) {
	deployments, err := c.client.AppsV1().Deployments("myproject").List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to load kubernetes configuration: %s\n", err)
	}

	components := make([]api.Component, len(deployments.Items))
	for i, item := range deployments.Items {
		components[i] = api.Component{
			Name: item.ObjectMeta.Name,
		}
	}

	return components, nil
}

func (c *cluster) WatchComponents() (internal.Observable, error) {
	watch, err := c.client.AppsV1().Deployments("myproject").Watch(metav1.ListOptions{})
	if err != nil {
		return internal.Observable{}, err
	}

	events := make(chan api.Event, 1)

	observable := internal.Observable{
		Events: events,
		Stop:   watch.Stop,
	}

	go func() {
		for event := range watch.ResultChan() {
			switch o := event.Object.(type) {
			case *appsv1.Deployment:
				fmt.Printf("Sending %v event for artifact %v in namespace\n", event.Type, o.Name)
				events <- api.Event{
					EventType: string(event.Type),
					Object: api.Component{
						Name: o.Name,
					},
				}
			default:
				fmt.Printf("unexpected event object of type %T\n", event.Object)
			}
		}
		fmt.Println("Stopped watching Deployments on Kubernetes cluster")
	}()

	return observable, nil
}
