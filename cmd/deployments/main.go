package main

import (
	// "k8s.io/client-go/kubernetes"

	"fmt"
	"log"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	// "k8s.io/client-go/tools/clientcmd/api"
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

func main() {
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	fmt.Println(config.Clusters[config.Default].CAData)

	clientcfg := &rest.Config{
		Host:        config.Clusters[config.Default].Host,
		BearerToken: config.Clusters[config.Default].Token,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: []byte(config.Clusters[config.Default].CAData),
		},
	}

	client, err := kubernetes.NewForConfig(clientcfg)
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	deploymentList, err := client.AppsV1().Deployments("myproject").List(v1.ListOptions{})
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	for _, deployment := range deploymentList.Items {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			fmt.Printf("%s (%s)\n", container.Name, container.Image)
		}
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
