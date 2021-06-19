package kubernetes

import (
	"database/sql"
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	apps "k8s.io/client-go/kubernetes/typed/apps/v1"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type DataSource interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type ConfigRepository struct {
	db DataSource
}

type Environment struct {
	id      string
	name    string
	configs map[string]*config
}

type config struct {
	host      string
	namespace string
	secret    string
	x509cert  []byte
}

type platform struct {
	namespace string
	client    kubernetes.Interface
}

var (
	cadata []byte
)

func init() {
	var err error
	cafile := "E:/projects/go/src/github.com/frohwerk/deputy-backend/certificates/minishift.crt"
	cadata, err = os.ReadFile(cafile)
	if err != nil {
		fmt.Printf("error reading cadata from %s: %s", cafile, err)
		os.Exit(1)
	}
}

func CreateConfigRepository(db DataSource) *ConfigRepository {
	return &ConfigRepository{db}
}

func (repo *ConfigRepository) Environment(envId string) (*Environment, error) {
	rows, err := repo.db.Query(`
	  SELECT envs.id, envs.name, platforms.name, platforms.api_server, platforms.namespace, platforms.secret
	    FROM envs JOIN platforms ON platforms.env_id = envs.id
	   WHERE envs.id = $1
	     AND platforms.api_server IS NOT NULL
		 AND platforms.namespace IS NOT NULL
		 AND platforms.secret IS NOT NULL
	`, envId)

	if err != nil {
		return nil, err
	}

	env := Environment{configs: map[string]*config{}}
	for i := 0; rows.Next(); i++ {
		var envId, envName, name, host, namespace, token string
		err := rows.Scan(&envId, &envName, &name, &host, &namespace, &token)

		if err != nil {
			return nil, err
		}

		if i == 0 {
			env.id = envId
			env.name = envName
		}

		env.configs[name] = &config{host: host, namespace: namespace, secret: token, x509cert: cadata}
	}

	return &env, nil
}

func (env *Environment) Platform(name string) (*platform, error) {
	if params, ok := env.configs[name]; ok {
		config := rest.Config{Host: params.host, BearerToken: params.secret, TLSClientConfig: rest.TLSClientConfig{CAData: params.x509cert}}

		client, err := kubernetes.NewForConfig(&config)
		if err != nil {
			return nil, err
		}

		return &platform{client: client, namespace: params.namespace}, nil
	} else {
		return nil, fmt.Errorf("configuration for platform %s is not available in environment %s", name, env.name)
	}
}

func (p *platform) Deployments() apps.DeploymentInterface {
	return p.client.AppsV1().Deployments(p.namespace)
}

func (p *platform) Pods() core.PodInterface {
	return p.client.Core().Pods(p.namespace)
}
