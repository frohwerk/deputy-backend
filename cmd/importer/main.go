package main

import (

	// "k8s.io/client-go/kubernetes"

	"bufio"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/frohwerk/deputy-backend/internal/database"
	yaml "gopkg.in/yaml.v2"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

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

	mutex      = &sync.Mutex{}
	yamlStdout = yaml.NewEncoder(os.Stdout)
	eventCount = 0
)

func main() {
	db = database.Open()
	defer database.Close(db)

	components = database.NewComponentStore(db)

	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	fmt.Println(config.Clusters[config.Default].CAData)

	client, err := kubernetes.NewForConfig(&rest.Config{
		Host:        config.Clusters[config.Default].Host,
		BearerToken: config.Clusters[config.Default].Token,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: []byte(config.Clusters[config.Default].CAData),
		},
	})
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	deploymentsWatch, err := client.AppsV1().Deployments("myproject").Watch(v1.ListOptions{})
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	podsWatch, err := client.CoreV1().Pods("myproject").Watch(v1.ListOptions{})
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	quit := make(chan interface{})
	deployments := deploymentsWatch.ResultChan()
	pods := podsWatch.ResultChan()

	go func() {
		r := bufio.NewReader(os.Stdin)
		if _, err := r.ReadString('\n'); err != nil {
			fmt.Printf("WARN  reading stdin failed: %s\n", err)
		}
		quit <- struct{}{}
	}()

	for {
		select {
		case <-quit:
			podsWatch.Stop()
			deploymentsWatch.Stop()
			return
		case event := <-deployments:
			handleEvent(event)
		case event := <-pods:
			handleEvent(event)
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
		if app, ok := o.ObjectMeta.Labels["app"]; ok && o.Status.Phase == core.PodRunning {
			imageid := o.Status.ContainerStatuses[0].ImageID
			if c, err := components.SetImage(app, strings.TrimPrefix(imageid, "docker-pullable://")); err != nil {
				log.Printf("ERROR Failed to update image for component %s: %s\n", app, err)
			} else {
				log.Printf("TRACE Updated image for component %s to %s\n", c.Name, c.Image)
			}
		} else {
			fmt.Printf("Pod %s has no app label, cannot find matching deployment\n", o.Name)
		}
		// for _, owner := range o.ObjectMeta.OwnerReferences {
		// 	if rs, err := client.AppsV1().ReplicaSets("myproject").Get(owner.Name, v1.GetOptions{}); err != nil {
		// 		log.Printf("ERROR failed to fetch ReplicaSet '%s': %s\n", owner.Name, err)
		// 	} else {
		// 		for _, owner := range rs.ObjectMeta.OwnerReferences {
		// 			owner.Name
		// 		}
		// 	}
		// }
		// fmt.Printf("Pod %s: %s (%s)\n", o.Name, event.Type, o.Status.Phase)
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

func logEvent(event watch.Event, o v1.ObjectMeta) {
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

func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configpath := filepath.Join(home, ".deputy")
	if _, err := os.Stat(configpath); os.IsNotExist(err) {
		return nil, fmt.Errorf("File $HOME/.deputy does not exist")
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
