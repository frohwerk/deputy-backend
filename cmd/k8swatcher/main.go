package main

import (

	// "k8s.io/client-go/kubernetes"

	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"

	"github.com/frohwerk/deputy-backend/internal/database"
	yaml "gopkg.in/yaml.v2"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	_ "github.com/lib/pq"
)

const (
	trace = false
)

type Config struct {
	Default  string             `yaml:"default"`
	Clusters map[string]Cluster `yaml:"clusters"`
}

type Cluster struct {
	Host    string `yaml:"host"`
	Account string `yaml:"account"`
	Token   string `yaml:"token"`
	CAData  string `yaml:"cadata"`
}

var (
	db         *sql.DB
	components database.ComponentStore

	client kubernetes.Interface

	mutex      = &sync.Mutex{}
	yamlStdout = yaml.NewEncoder(os.Stdout)
	eventCount = 0
)

func main() {
	db = database.Open()
	defer database.Close(db)

	components = database.NewComponentStore(db)

	kubeclient, err := LoadKubeconfig()
	// client, err := LoadClient()
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	client = kubeclient

	deploymentsWatch, err := client.AppsV1().Deployments("myproject").Watch(metav1.ListOptions{})
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	deployments := deploymentsWatch.ResultChan()

	// podsWatch, err := client.CoreV1().Pods("myproject").Watch(metav1.ListOptions{})
	// if err != nil {
	// 	log.Fatalf("%s\n", err)
	// }
	// pods := podsWatch.ResultChan()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	for {
		select {
		case sig := <-signals:
			switch sig {
			case os.Interrupt:
				fmt.Println("Received SIGINT")
				deploymentsWatch.Stop()
				return
				// podsWatch.Stop()
			case os.Kill:
				fmt.Println("Received SIGTERM")
				deploymentsWatch.Stop()
				return
				// podsWatch.Stop()
			default:
				fmt.Fprintf(os.Stderr, "Received unexpected signal: %v\n", sig)
			}
		case event := <-deployments:
			handleEvent(event)
			// case event := <-pods:
			// 	handleEvent(event)
		}
	}
}

func handleEvent(event watch.Event) {
	switch o := event.Object.(type) {
	case *apps.Deployment:
		if event.Type == watch.Added {
			if c, err := components.CreateIfAbsent(o.Name); err != nil {
				log.Printf("ERROR Failed to register component '%s': %s\n", o.Name, err)
			} else {
				log.Printf("TRACE Component '%s' is registered with id '%s'\n", o.Name, c.Id)
			}
		}
		imageid := o.Spec.Template.Spec.Containers[0].Image
		if c, err := components.SetImage(o.Name, strings.TrimPrefix(imageid, "docker-pullable://")); err != nil {
			log.Printf("ERROR Failed to update image for component %s: %s\n", o.Name, err)
		} else {
			log.Printf("TRACE Updated image for component %s to %s\n", c.Name, c.Image)
		}
		if trace {
			logEvent(event, o.ObjectMeta)
		}
		// fmt.Printf("Deployment %s: %s\n", o.Name, event.Type)
	case *core.Pod:
		// TODO: Find matching deployment
		// TODO: Support different labels
		if trace {
			logEvent(event, o.ObjectMeta)
		}
		if o.Status.Phase == core.PodRunning {
			name, err := GetName(&o.ObjectMeta)
			if err != nil {
				fmt.Printf("error finding name for pod %s: %s\n", o.Name, err)
				return
			}
			if c, err := components.CreateIfAbsent(name); err != nil {
				log.Printf("ERROR Failed to register component '%s': %s\n", o.Name, err)
			} else {
				log.Printf("TRACE Component '%s' is registered with id '%s'\n", c.Name, c.Id)
			}
			imageid := o.Status.ContainerStatuses[0].ImageID
			if c, err := components.SetImage(name, strings.TrimPrefix(imageid, "docker-pullable://")); err != nil {
				log.Printf("ERROR Failed to update image for component %s: %s\n", name, err)
			} else {
				log.Printf("TRACE Updated image for component %s to %s\n", c.Name, c.Image)
			}
			// TODO: scan for matches with registered artifacts
			// If one artifact matches with multiple versions, use the most precise match (most matched files)
		}
	default:
		printYaml(event)
	}
}

func filename(objectType string, objectName string) string {
	mutex.Lock()
	defer mutex.Unlock()
	eventCount++
	return fmt.Sprintf(`temp/logs/%04d-%s-%s.yaml`, eventCount, objectType, objectName)
}

func logEvent(event watch.Event, o metav1.ObjectMeta) {
	kind := ""
	switch event.Object.(type) {
	case *apps.Deployment:
		kind = "deployment"
	case *core.Pod:
		kind = "pod"
	}
	if f, err := os.Create(filename(kind, o.Name)); err != nil {
		fmt.Printf("ERROR failed to create file %s: %s\n", f.Name(), err)
	} else {
		defer func() {
			if err := f.Close(); err != nil {
				fmt.Printf("ERROR failed to close file %s: %s\n", f.Name(), err)
			}
		}()
		writeYaml(event, f)
	}
}

func LoadKubeconfig() (kubernetes.Interface, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configpath := filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", configpath)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func LoadClient() (kubernetes.Interface, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	fmt.Println(config.Clusters[config.Default].CAData)

	return kubernetes.NewForConfig(&rest.Config{
		Host:        config.Clusters[config.Default].Host,
		BearerToken: config.Clusters[config.Default].Token,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: []byte(config.Clusters[config.Default].CAData),
		},
	})
}

func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// configpath := filepath.Join(home, ".kube", "config")
	configpath := filepath.Join(home, ".deputy")
	if _, err := os.Stat(configpath); os.IsNotExist(err) {
		return nil, fmt.Errorf("File $%s does not exist", configpath)
	}

	configfile, err := os.Open(configpath)
	if err != nil {
		return nil, fmt.Errorf("Error opening file $HOME/.deputy: %s", err)
	}

	config := new(Config)
	if err := yaml.NewDecoder(configfile).Decode(config); err != nil {
		return nil, fmt.Errorf("Error reading file $HOME/.deputy: %s", err)
	}

	// for name, cluster := range config.Clusters {
	// 	if cadata, err := base64.StdEncoding.DecodeString(config.Clusters[config.Default].CAData); err != nil {
	// 		return nil, fmt.Errorf("Error reading attribute CAData for cluster %s: %s", name, err)
	// 	} else {
	// 		log.Println("Decoded cadata:")
	// 		log.Println(cadata)
	// 		cluster.CAData = string(cadata)
	// 	}
	// }

	return config, nil
}

func writeYaml(e watch.Event, w io.Writer) {
	encoder := yaml.NewEncoder(w)
	defer encoder.Close()
	if err := encoder.Encode(e); err != nil {
		fmt.Printf("ERROR decoding of watch event failed: %s\n", err)
	}
}

func printYaml(e watch.Event) {
	if err := yamlStdout.Encode(e); err != nil {
		fmt.Printf("ERROR decoding of watch event failed: %s\n", err)
	}
}

func GetName(o *metav1.ObjectMeta) (string, error) {
	for _, owner := range o.OwnerReferences {
		switch owner.Kind {
		case "ReplicaSet":
			rs, err := client.AppsV1().ReplicaSets(o.Namespace).Get(owner.Name, metav1.GetOptions{})
			if err != nil {
				fmt.Println("failed to fetch ReplicaSet", o.Name)
				fmt.Println(err)
				continue
			}
			return GetName(&rs.ObjectMeta)
		case "Deployment":
			return owner.Name, nil
		default:
			fmt.Println("Unsupported owner type", owner.Kind)
			continue
		}
	}
	return o.Name, nil
}
