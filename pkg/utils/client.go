package utils

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

func Cli() *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/weibh/.kube/config")
	if err != nil {
		log.Fatal(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err.Error())
	}
	return clientset
}
