package main

import (
	"flag"
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	master     = flag.String("master", "", "URL of the Kubernetes API server.")
	kubeconfig = flag.String("kubeconfig", "", "Path to a kubeconfig.")
)

func main() {
	flag.Parse()
	log.Printf("yolo up")
	cfg, err := clientcmd.BuildConfigFromFlags(*master, *kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}
	_, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %v", err)
	}
}
