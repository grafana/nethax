package pkg

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
)

type Kubernetes struct {
	config *rest.Config
	client *kubernetes.Clientset
}

// Fetch .kube/config file or generate it from a flag
func getKubeConfig() *rest.Config {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")

	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")

	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())

	}

	return config

}

func makeKubeClient(config *rest.Config) *kubernetes.Clientset {
	// create the clientset
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return client

}

func initializeKubeClient() *Kubernetes {

	config := getKubeConfig()
	client := makeKubeClient(config)

	return &Kubernetes{config, client}
}

func getNamespaces(k *Kubernetes) []string {
	namespaces, err := k.client.CoreV1().Namespaces().List(
		context.TODO(),
		metav1.ListOptions{})

	if err != nil {
		panic(err.Error())

	}
	nsNames := []string{}
	for _, ns := range namespaces.Items {
		nsNames = append(nsNames, ns.Name)

	}
	return nsNames

}

func getPods(k *Kubernetes, namespace string) []string {
	pods, err := k.client.CoreV1().Pods(namespace).List(
		context.TODO(),
		metav1.ListOptions{})

	if err != nil {
		panic(err.Error())

	}
	podNames := []string{}
	for _, pod := range pods.Items {
		podNames = append(podNames, pod.Name)

	}
	return podNames

}

func createContainer(k *Kubernetes, namespace string, pod string) {
	containerName := "nethack"
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		pod, err := k.client.CoreV1().Pods(namespace).Get(
			context.TODO(),
			pod,
			metav1.GetOptions{})
		if err != nil {
			return err
		}

		// Check if the container already exists
		for _, c := range pod.Spec.Containers {
			if c.Name == containerName {
				return fmt.Errorf("container %s already exists in pod %s",
					containerName,
					pod)
			}
		}

		// Add the new container to the pod
		pod.Spec.Containers = append(pod.Spec.Containers,
			corev1.Container{
				Name:  containerName,
				Image: "nicolaka/netshoot",
			})

		_, err = k.client.CoreV1().Pods(namespace).Update(
			context.TODO(),
			pod,
			metav1.UpdateOptions{})
		return err
	})
	if err != nil {
		log.Fatalf("Error creating container in pod: %v", err)
	}

}

func pollContainer(k *Kubernetes, namespace string, pod string, container string) {
	// Wait for the new container to start running
}
