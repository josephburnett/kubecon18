package main

import (
	"flag"
	"log"
	"time"

	"github.com/josephburnett/kubecon-seattle-2018/yolo/pkg/apis/autoscaling/v1alpha1"
	clientset "github.com/josephburnett/kubecon-seattle-2018/yolo/pkg/client/clientset/versioned"
	duck "github.com/knative/pkg/apis/duck/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

// Pro Tip: if you want to do this in production, copy
// https://github.com/kubernetes/sample-controller which makes use of
// client-go informers and implements controller best practices.

func main() {

	w := establishWatch()
	resync := time.NewTicker(30 * time.Second).C

	for {
		select {
		case event, ok := <-w.ResultChan():
			if !ok {
				w.Stop()
				w = establishWatch()
				continue
			}
			pa := event.Object.(*v1alpha1.PodAutoscaler)
			switch event.Type {
			case watch.Added:

				// Take control of yolo-class PodAutoscalers only
				if pa.Annotations["autoscaling.knative.dev/class"] == "yolo" {

					// Calculate a recommended scale
					replicas := recommendedScale(pa)

					// Update the Deployment
					scaleTo(pa, replicas)

					// Update status
					updateStatus(pa, replicas)

				}
			}
		case <-resync:
			w.Stop()
			w = establishWatch()
		}
	}
}

func recommendedScale(pa *v1alpha1.PodAutoscaler) int32 {

	// Do something really smart here ...
	return 2
}

func scaleTo(pa *v1alpha1.PodAutoscaler, replicas int32) {
	deployment, err := kubeClient.AppsV1().Deployments("kubecon-seattle-2018").Get(pa.Name+"-deployment", v1.GetOptions{})
	if err != nil {
		log.Printf("Error getting Deployment %q: %v", pa.Name, err)
		return
	}
	deployment.Spec.Replicas = &replicas
	if _, err := kubeClient.AppsV1().Deployments("kubecon-seattle-2018").Update(deployment); err != nil {
		log.Printf("Error updating Deployment %q: %v", pa.Name, err)
	}
}

func updateStatus(oldPa *v1alpha1.PodAutoscaler, replicas int32) {
	pa, err := autoscalingClient.AutoscalingV1alpha1().PodAutoscalers("kubecon-seattle-2018").Get(oldPa.Name, v1.GetOptions{})
	if err != nil {
		log.Printf("Error getting PodAutoscaler %q: %v.", pa.Name, err)
		return
	}
	// pa.Status.RecommendedScale = &replicas
	pa.Status.MarkActive()
	pa.Status.SetConditions(duck.Conditions{{
		Type:   duck.ConditionReady,
		Status: corev1.ConditionTrue,
		Reason: "I was born ready",
	}})
	if _, err := autoscalingClient.AutoscalingV1alpha1().PodAutoscalers("kubecon-seattle-2018").Update(pa); err != nil {
		log.Printf("Error updating PodAutoscaler %q: %v.", pa.Name, err)
	}
}

var (
	kubeconfig        string
	masterURL         string
	kubeClient        kubernetes.Interface
	autoscalingClient clientset.Interface
)

func init() {

	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")

	// Make Kubernetes and Knative Autoscaling clients.
	flag.Parse()
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}
	kubeClient, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %v", err)
	}
	autoscalingClient, err = clientset.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building autoscaling clientset: %v", err)
	}
}

func establishWatch() watch.Interface {
	// Watch for changes
	w, err := autoscalingClient.AutoscalingV1alpha1().PodAutoscalers("kubecon-seattle-2018").Watch(v1.ListOptions{})
	if err != nil {
		panic(err)
	}
	return w
}
