package main

import (

	// "k8s.io/client-go/kubernetes"

	"database/sql"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"

	"github.com/frohwerk/deputy-backend/internal/database"
	"github.com/frohwerk/deputy-backend/internal/logger"
	"github.com/frohwerk/deputy-backend/pkg/api"
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

type eventHandler *sql.DB

var (
	platform *api.Platform

	db         *sql.DB
	components database.ComponentStore

	client kubernetes.Interface

	log logger.Logger = logger.Basic(logger.LEVEL_DEBUG)

	mutex      = &sync.Mutex{}
	yamlStdout = yaml.NewEncoder(os.Stdout)
	eventCount = 0
)

func main() {
	// minishift = c49ca75c-da18-4641-950c-f5609877828f
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <platform>", os.Args[0])
		fmt.Printf("error: No target platform specified")
		os.Exit(1)
	}

	for _, todo := range todos {
		fmt.Println("TODO:", todo)
	}

	db = database.Open()
	defer database.Close(db)

	platforms := database.NewPlatformStore(db)
	components = database.NewComponentStore(db)
	deployments := database.NewDeploymentStore(db)

	p, err := platforms.Get(os.Args[1])
	if err != nil {
		fmt.Printf("error reading platform configuration from database: %s", err)
		os.Exit(1)
	}
	platform = p

	fmt.Println("TODO: replace hard coded certificate file")
	fmt.Println("TODO: add option to get secret and certificate file from a file reference (useful fur cluster deployment)")
	cafile := "E:/projects/go/src/github.com/frohwerk/deputy-backend/certificates/minishift.crt"
	cadata, err := os.ReadFile(cafile)
	if err != nil {
		fmt.Printf("error reading cadata from %s: %s", cafile, err)
		os.Exit(1)
	}

	config := &rest.Config{
		Host:        platform.ServerUri,
		BearerToken: platform.Secret,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: cadata,
		},
	}

	if kubeclient, err := kubernetes.NewForConfig(config); err != nil {
		fmt.Printf("error initializing kubernetes client: %s", err)
		os.Exit(1)
	} else {
		client = kubeclient
	}

	deploymentsWatch, err := client.AppsV1().Deployments(platform.Namespace).Watch(metav1.ListOptions{})
	if err != nil {
		fmt.Printf("error watching namespace %s on api-server %s: %s\n", platform.Namespace, platform.ServerUri, err)
		os.Exit(1)
	}

	shutdown := sync.WaitGroup{}
	shutdown.Add(1)

	go func() {
		defer shutdown.Done()
		handler := &deploymentHandler{platform, components, deployments}
		for event := range deploymentsWatch.ResultChan() {
			err := handler.handleEvent(event)
			if err != nil {
				log.Error("error handling deployment event: %s", err)
			}
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	for {
		switch sig := <-signals; sig {
		case os.Interrupt:
			fmt.Println("Received SIGINT")
			deploymentsWatch.Stop()
			shutdown.Wait()
			return
		case os.Kill:
			fmt.Println("Received SIGTERM")
			deploymentsWatch.Stop()
			return
		default:
			fmt.Fprintf(os.Stderr, "Received unexpected signal: %v\n", sig)
		}
	}
}

func handleEvent(event watch.Event) {
	switch o := event.Object.(type) {
	case *apps.Deployment:
		fmt.Println("TODO  Use transaction for multiple database changes")
		if event.Type == watch.Added {
			if c, err := components.CreateIfAbsent(o.Name); err != nil {
				fmt.Printf("ERROR Failed to register component '%s': %s\n", o.Name, err)
			} else {
				fmt.Printf("TRACE Component '%s' is registered with id '%s'\n", o.Name, c.Id)
			}
		}
		fmt.Println("TODO: write deployments entry")
		c, err := components.SetImage(o.Name, strings.TrimPrefix(o.Spec.Template.Spec.Containers[0].Image, "docker-pullable://"))
		if err != nil {
			fmt.Printf("ERROR Failed to update image for component %s: %s\n", o.Name, err)
			return
		} else {
			fmt.Printf("TRACE Updated image for component %s to %s\n", c.Name, c.Image)
		}
		if err := components.LinkPlatform(platform.Id, c.Id); err != nil {
			fmt.Printf("ERROR Failed to link platform %s with image %s: %s\n", platform.Id, c.Id, err)
			return
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
				fmt.Printf("ERROR Failed to register component '%s': %s\n", o.Name, err)
			} else {
				fmt.Printf("TRACE Component '%s' is registered with id '%s'\n", c.Name, c.Id)
			}
			imageid := o.Status.ContainerStatuses[0].ImageID
			if c, err := components.SetImage(name, strings.TrimPrefix(imageid, "docker-pullable://")); err != nil {
				fmt.Printf("ERROR Failed to update image for component %s: %s\n", name, err)
			} else {
				fmt.Printf("TRACE Updated image for component %s to %s\n", c.Name, c.Image)
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
