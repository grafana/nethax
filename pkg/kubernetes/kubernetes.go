package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	// ProbeImageVersion is set at build time via ldflags
	DefaultProbeImage = "grafana/nethax-probe:latest"
)

// New returns a new Kubernetes object, connected to the given
// context, or to the in-cluster API if blank.
func New(context string) (*Kubernetes, error) {
	config, err := getClusterConfig(context)
	if err != nil {
		return nil, fmt.Errorf("fetching Kubernetes configuration: %w", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating Kubernetes client: %w", err)
	}

	return &Kubernetes{
		client: client,
	}, nil
}

func getClusterConfig(kontext string) (*rest.Config, error) {
	// attempt to use config from pod service account
	cfg, err := rest.InClusterConfig()
	if err != nil {
		// Can be overridden by KUBECONFIG variable
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverride := &clientcmd.ConfigOverrides{
			CurrentContext: kontext,
		}

		cfg, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			loadingRules,
			configOverride,
		).ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("loading Kubernetes configuration: %w", err)
		}
	}

	return cfg, nil
}

type Kubernetes struct {
	client kubernetes.Interface
}

var (
	errNoPodsFound = errors.New("no pods found")
)

func (k *Kubernetes) GetPods(ctx context.Context, namespace, labels, fields string) ([]corev1.Pod, error) {
	pods, err := k.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels,
		FieldSelector: fields,
	})
	if err != nil {
		return nil, fmt.Errorf("listing pods for namespace %s, labels %q, and fields %q: %w", namespace, labels, fields, err)
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("%w: namespace %s, labels %q, fields %q", errNoPodsFound, namespace, labels, fields)
	}

	return pods.Items, nil
}

func GetProbeImage(probeImage string) string {
	// Use the provided probe image, or default if empty
	if probeImage == "" {
		probeImage = DefaultProbeImage
	}
	return probeImage
}

func (k *Kubernetes) LaunchEphemeralContainer(ctx context.Context, pod *corev1.Pod, probeImage string, command []string, args []string) (*corev1.Pod, string, error) {
	podJS, err := json.Marshal(pod)
	if err != nil {
		return nil, "", fmt.Errorf("error creating JSON for pod: %v", err)
	}

	ephemeralName := fmt.Sprintf("nethax-probe-%v", time.Now().UnixNano())

	debugContainer := &corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:    ephemeralName,
			Image:   GetProbeImage(probeImage),
			Command: command,
			Args:    args,
		},
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

	pods := k.client.CoreV1().Pods(pod.Namespace)
	result, err := pods.Patch(ctx, pod.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{}, "ephemeralcontainers")
	if err != nil {
		return nil, ephemeralName, fmt.Errorf("error patching pod with debug container: %v", err)
	}

	return result, ephemeralName, nil
}

var errEphemeralContainerNotFound = errors.New("ephemeral container not found")

func (k *Kubernetes) PollEphemeralContainerStatus(ctx context.Context, pod *corev1.Pod, ephemeralContainerName string) (int32, error) {
	interval, timeout := time.Second, 30*time.Second // TODO(inkel) make these arguments

	var state corev1.ContainerState

	err := wait.PollUntilContextTimeout(ctx, interval, timeout, false, func(ctx context.Context) (bool, error) {
		pod, err := k.client.CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("getting pod: %w", err)
		}

		for _, s := range pod.Status.EphemeralContainerStatuses {
			if s.Name == ephemeralContainerName {
				state = s.State
				return s.State.Terminated != nil, nil
			}
		}

		return false, errEphemeralContainerNotFound
	})
	if err != nil {
		return -1, fmt.Errorf("polling ephemeral container terminated state: %w", err)
	}

	return state.Terminated.ExitCode, nil
}
