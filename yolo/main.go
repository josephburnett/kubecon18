package main

import (
	"flag"
	"log"

	clientset "github.com/knative/serving/pkg/client/clientset/versioned"
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
	kubeClient, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %v", err)
	}
	knativeClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building serving clientset: %v", err)
	}
	_, err = kubeClient.CoreV1().Pods("").Watch(metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error establishing watch on pods: %v", err)
	}
	_, err = knativeClient.Autoscaling().V1alpha1().PodAutoscalers("").Watch(metav1.ListOptions())
	if err != nil {
		log.Fatalf("Error establishing watch on PodAutoscalers: %v", err)
	}
}
