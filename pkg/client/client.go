package client

import (
	"log"
	"os"

	"k8s.io/client-go/rest"
)

type ClientSet struct{}

type Clients interface {
	GetConfig() *rest.Config
}

func (c *ClientSet) GetConfig() *rest.Config {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Print("Failed to instantiate k8s client: ", err)
		os.Exit(1)
	}
	return config
}

func Init() Clients {
	return &ClientSet{}
}
