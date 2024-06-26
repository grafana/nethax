package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	instance *Kubernetes = &Kubernetes{}
)

type Kubernetes struct {
	Config *rest.Config
	Client kubernetes.Interface
}

func fetchKubeConfig() {
	// attempt to use config from pod service account
	config, err := rest.InClusterConfig()
	if err != nil {
		// TODO - allow overriding of kubeconfig path
		// use the current context in kubeconfig -- assume it is in the home dir
		config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
		if err != nil {
			panic(err.Error())
		}
	}

	instance.Config = config
}

func makeKubeClient() {
	client, err := kubernetes.NewForConfig(instance.Config)
	if err != nil {
		panic(err.Error())
	}

	instance.Client = client
}

func GetKubernetes() *Kubernetes {
	if instance.Config == nil {
		fetchKubeConfig()
	}
	if instance.Client == nil {
		makeKubeClient()
	}

	// you are now ready to Kubernetes.
	return instance
}

func GetPods(namespace string) []string {
	k := GetKubernetes()
	pods, err := k.Client.CoreV1().Pods(namespace).List(
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

func chooseTargetContainer(pod *corev1.Pod) string {
	// TODO add capability to pick container by name (currently assume 0th container)
	if len(pod.Spec.Containers) == 0 {
		log.Fatalf("Error: No containers in pod.")
	}
	return pod.Spec.Containers[0].Name
}

func LaunchEphemeralContainer(pod *corev1.Pod, command []string, args []string) (*corev1.Pod, string, error) {
	k := GetKubernetes()
	podJS, err := json.Marshal(pod)
	if err != nil {
		return nil, "", fmt.Errorf("error creating JSON for pod: %v", err)
	}

	ephemeralName := fmt.Sprintf("nethax-netshoot-%v", time.Now().UnixNano())

	debugContainer := &corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:    ephemeralName,
			Image:   "nicolaka/netshoot",
			Command: command,
			Args:    args,
		},
		TargetContainerName: chooseTargetContainer(pod),
	}

	debugPod := pod.DeepCopy()
	debugPod.Spec.EphemeralContainers = append(debugPod.Spec.EphemeralContainers, *debugContainer)

	debugJS, err := json.Marshal(debugPod)
	if err != nil {
		return nil, ephemeralName, fmt.Errorf("error creating JSON for debug container: %v", err)
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(podJS, debugJS, pod)
	if err != nil {
		return nil, ephemeralName, fmt.Errorf("error creating patch to add debug container: %v", err)
	}

	pods := k.Client.CoreV1().Pods(pod.Namespace)
	result, err := pods.Patch(context.TODO(), pod.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{}, "ephemeralcontainers")

	return result, ephemeralName, nil
}

func getEphemeralContainerExitStatus(pod *corev1.Pod, ephemeralContainerName string) (int32, error) {
	k := GetKubernetes()
	pod, err := k.Client.CoreV1().Pods(pod.Namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
	if err != nil {
		return -1, err
	}

	for _, ec := range pod.Status.EphemeralContainerStatuses {
		if ec.Name == ephemeralContainerName {
			if ec.State.Terminated != nil && ec.State.Terminated.ExitCode > -1 {
				return ec.State.Terminated.ExitCode, nil
			}
		}
	}
	return -1, nil
}

func isEphemeralContainerTerminated(pod *corev1.Pod, ephemeralContainerName string) wait.ConditionWithContextFunc {
	return func(ctx context.Context) (bool, error) {
		exitCode, err := getEphemeralContainerExitStatus(pod, ephemeralContainerName)
		if err != nil {
			return false, err
		}
		if exitCode > -1 {
			return true, nil
		}
		return false, nil
	}
}

// Poll up to timeout seconds for pod to enter running state.
// Returns an error if the pod never enters the running state.
func waitForEphemeralContainerTerminated(pod *corev1.Pod, ephemeralContainerName string, timeout time.Duration) error {
	return wait.PollUntilContextTimeout(context.TODO(), time.Second, timeout, false, isEphemeralContainerTerminated(pod, ephemeralContainerName))
}

func PollEphemeralContainerStatus(pod *corev1.Pod, ephemeralContainerName string) int32 {
	// poll until ephemeral container has an exit status
	err := waitForEphemeralContainerTerminated(pod, ephemeralContainerName, time.Second*30)
	if err != nil {
		fmt.Println("Error waiting for ephemeral container start.", err)
		os.Exit(2)
	}
	// return exit status
	exitCode, err := getEphemeralContainerExitStatus(pod, ephemeralContainerName)
	if err != nil {
		fmt.Println("Error getting ephemeral container exit code.", err)
		os.Exit(3)
	}

	return exitCode
}
