package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/grafana/nethax/pkg/common"
)

var (
	// ProbeImageVersion is set at build time via ldflags
	ProbeImageVersion = "latest"
)

var (
	// Singleton for Kubernetes access
	instance *Kubernetes = &Kubernetes{}
)

func makeKubeClient() {
	client, err := kubernetes.NewForConfig(instance.Config)
	if err != nil {
		panic(err.Error())
	}

	instance.Client = client
}

func GetKubernetes(context string) *Kubernetes {
	if instance.Config == nil {
		fetchKubeConfig(context)
	}
	if instance.Client == nil {
		makeKubeClient()
	}

	// you are now ready to Kubernetes.
	return instance
}

func fetchKubeConfig(context string) {
	// attempt to use config from pod service account
	config, err := rest.InClusterConfig()
	if err != nil {
		// Can be overridden by KUBECONFIG variable
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverride := &clientcmd.ConfigOverrides{}
		if context != "" {
			configOverride.CurrentContext = context
		}

		config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			loadingRules,
			configOverride,
		).ClientConfig()

		if err != nil {
			panic(err.Error())
		}
	}

	instance.Config = config
}

type Kubernetes struct {
	Config *rest.Config
	Client kubernetes.Interface
}

func (k *Kubernetes) GetPods(ctx context.Context, namespace string) []string {
	pods, err := k.Client.CoreV1().Pods(namespace).List(
		ctx,
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

func (k *Kubernetes) LaunchEphemeralContainer(ctx context.Context, pod *corev1.Pod, command []string, args []string) (*corev1.Pod, string, error) {
	podJS, err := json.Marshal(pod)
	if err != nil {
		return nil, "", fmt.Errorf("error creating JSON for pod: %v", err)
	}

	ephemeralName := fmt.Sprintf("nethax-probe-%v", time.Now().UnixNano())

	debugContainer := &corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:    ephemeralName,
			Image:   fmt.Sprintf("nethax-probe:%s", ProbeImageVersion),
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
	result, err := pods.Patch(ctx, pod.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{}, "ephemeralcontainers")
	if err != nil {
		return nil, ephemeralName, fmt.Errorf("error patching pod with debug container: %v", err)
	}

	return result, ephemeralName, nil
}

func (k *Kubernetes) getEphemeralContainerExitStatus(ctx context.Context, pod *corev1.Pod, ephemeralContainerName string) (int32, error) {
	pod, err := k.Client.CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
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

func (k *Kubernetes) isEphemeralContainerTerminated(pod *corev1.Pod, ephemeralContainerName string) wait.ConditionWithContextFunc {
	return func(ctx context.Context) (bool, error) {
		exitCode, err := k.getEphemeralContainerExitStatus(ctx, pod, ephemeralContainerName)
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
func (k *Kubernetes) waitForEphemeralContainerTerminated(ctx context.Context, pod *corev1.Pod, ephemeralContainerName string, timeout time.Duration) error {
	return wait.PollUntilContextTimeout(ctx, time.Second, timeout, false, k.isEphemeralContainerTerminated(pod, ephemeralContainerName))
}

func (k *Kubernetes) PollEphemeralContainerStatus(ctx context.Context, pod *corev1.Pod, ephemeralContainerName string) int32 {
	// poll until ephemeral container has an exit status
	err := k.waitForEphemeralContainerTerminated(ctx, pod, ephemeralContainerName, time.Second*30)
	if err != nil {
		fmt.Println("Error waiting for ephemeral container start.", err)
		common.ExitNethaxError()
	}
	// return exit status
	exitCode, err := k.getEphemeralContainerExitStatus(ctx, pod, ephemeralContainerName)
	if err != nil {
		fmt.Println("Error getting ephemeral container exit code.", err)
		common.ExitNethaxError()
	}

	return exitCode
}
