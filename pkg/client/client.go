package client

import (
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type ClientSet struct {
	ClientSet     *kubernetes.Clientset
	DynamicClient *dynamic.DynamicClient
}

type Clients interface {
	Client() (*kubernetes.Clientset, *dynamic.DynamicClient)
}

func (C ClientSet) Client() (*kubernetes.Clientset, *dynamic.DynamicClient) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("error in getting user home dir: %v\n", err)
		os.Exit(1)
	}

	// get kubeconfig path
	kubeConfigPath := filepath.Join(homedir, ".kube", "config")
	log.Printf("KubeConfig file path is: %v\n", kubeConfigPath)

	// build kubeconfig file from path
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		log.Printf("error in loading kubeconfig file: %v\n", err)
		os.Exit(1)
	}

	// creates new kubernetes client
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		log.Printf("error in creating kuberentes client: %v\n", err)
		os.Exit(1)
	}
	dynamicLient, err := dynamic.NewForConfig(kubeconfig)
	if err != nil {
		panic(err)
	}
	return clientset, dynamicLient
}

func InitClient() Clients {
	return &ClientSet{}
}
