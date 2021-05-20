package kubernetes

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Config struct {
	ServerUri string
	Secret    string
	CAData    string
}

func NewClient(server, secret string, cadata []byte) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(
		&rest.Config{
			Host:            server,
			BearerToken:     secret,
			TLSClientConfig: rest.TLSClientConfig{CAData: cadata},
		},
	)
}
