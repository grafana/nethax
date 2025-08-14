package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"
	"testing/synctest"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	testClient "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func setup() *Kubernetes {
	return &Kubernetes{
		client: testClient.NewSimpleClientset(),
	}
}

func TestGetPods(t *testing.T) {
	k := &Kubernetes{
		client: testClient.NewSimpleClientset(
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "nethax",
					Name:      "example-pod-001",
					Labels: map[string]string{
						"app": "nethax",
					},
				},
				Spec: corev1.PodSpec{
					NodeName: "foo-bar-42",
				},
			},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "grafana",
					Name:      "example-pod-002",
					Labels: map[string]string{
						"app": "grafana",
					},
				},
				Spec: corev1.PodSpec{
					NodeName: "foo-bar-42",
				},
			},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "nethax",
					Name:      "example-pod-002",
					Labels: map[string]string{
						"app": "nethax",
						"foo": "bar",
					},
				},
				Spec: corev1.PodSpec{
					NodeName: "lorem-ipsum-23",
				},
			},
		),
	}

	t.Run("not found", func(t *testing.T) {
		tests := []struct {
			ns, labels string
		}{
			{corev1.NamespaceAll, "app=redis"},
			{"nethax", "app=postgresql"},
			{"grafana", "app=nethax"},
			{"mimir", ""},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("ns=%s,sel=%s", tt.ns, tt.labels), func(t *testing.T) {
				pods, err := k.GetPods(t.Context(), tt.ns, tt.labels, "")
				if !errors.Is(err, errNoPodsFound) {
					t.Fatalf("expecting error %v, got %v", errNoPodsFound, err)
				}

				if len(pods) != 0 {
					t.Fatalf("expecting no pods to be returned, got %v", pods)
				}
			})
		}
	})

	t.Run("found", func(t *testing.T) {
		tests := []struct {
			ns, labels string
			exp        int
		}{
			{corev1.NamespaceAll, "", 3},
			{"nethax", "", 2},
			{"nethax", "foo=bar", 1},
		}

		for _, tt := range tests {
			pods, err := k.GetPods(t.Context(), tt.ns, tt.labels, "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got := len(pods); tt.exp != got {
				t.Fatalf("ns %q, labels %q: expecting %d pods, got %d", tt.ns, tt.labels, tt.exp, got)
			}
		}
	})
}

func TestLaunchEphemeralContainer(t *testing.T) {

	// prepare
	k := setup()
	name := "grafanyaa"
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: name,
			Name:      name,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "owo"},
			},
		},
	}
	k.client.CoreV1().Pods(name).Create(t.Context(), pod, metav1.CreateOptions{})

	// execute
	actual, _, _ := k.LaunchEphemeralContainer(t.Context(), pod, "", []string{"nyaa"}, []string{"rawr"})

	// assert
	if len(actual.Spec.EphemeralContainers) != 1 {
		t.Fatalf("Expected PodSpec EphemeralContainers to be 1, got: %d", len(actual.Spec.EphemeralContainers))
	}
}

func TestLaunchEphemeralContainerWithProbeImage(t *testing.T) {
	tests := []struct {
		name          string
		probeImage    string
		expectedImage string
	}{
		{
			name:          "default probe image when empty",
			probeImage:    "",
			expectedImage: DefaultProbeImage,
		},
		{
			name:          "custom probe image",
			probeImage:    "grafanyaa/mangodb-probe:v1.2.3",
			expectedImage: "grafanyaa/mangodb-probe:v1.2.3",
		},
		{
			name:          "fully qualified image",
			probeImage:    "gcr.io/project/nethax-probe:abc123",
			expectedImage: "gcr.io/project/nethax-probe:abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := setup()

			// Create a minimal pod just for testing
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-ns",
					Name:      "test-pod",
				},
			}
			k.client.CoreV1().Pods("test-ns").Create(t.Context(), pod, metav1.CreateOptions{})

			// Execute with specific probe image
			result, _, err := k.LaunchEphemeralContainer(t.Context(), pod, tt.probeImage, []string{"test"}, []string{"arg"})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify only what we care about - the image
			if len(result.Spec.EphemeralContainers) != 1 {
				t.Fatalf("Expected 1 ephemeral container, got: %d", len(result.Spec.EphemeralContainers))
			}

			if result.Spec.EphemeralContainers[0].Image != tt.expectedImage {
				t.Errorf("Expected image %s, got %s", tt.expectedImage, result.Spec.EphemeralContainers[0].Image)
			}
		})
	}
}

func TestPollEphemeralContainerStatus(t *testing.T) {
	const (
		ns      = "foo"
		podName = "bar"

		ephemeralContainer = "quux"
	)

	t.Run("success", func(t *testing.T) {
		exitCode := rand.Int32N(128)

		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      podName,
			},
			Status: corev1.PodStatus{
				EphemeralContainerStatuses: []corev1.ContainerStatus{
					{
						Name:  ephemeralContainer,
						State: corev1.ContainerState{},
					},
				},
			},
		}

		var stage int

		c := testClient.NewClientset()

		// This reactor is a function that gets executed every time we
		// call the get pods API. It has different stages to mimick
		// waiting for the ephemeral container to finish.
		//
		// ktesting "k8s.io/client-go/testing"
		c.PrependReactor("get", "pods", func(_ ktesting.Action) (bool, runtime.Object, error) {
			var state corev1.ContainerState

			switch stage {
			case 0: // initializing
				state.Waiting = &corev1.ContainerStateWaiting{}
			case 1: // running
				state.Running = &corev1.ContainerStateRunning{
					StartedAt: metav1.Now(),
				}
			case 2: // do work
				state = pod.Status.EphemeralContainerStatuses[0].State
			case 3: // terminated
				state.Terminated = &corev1.ContainerStateTerminated{
					ExitCode: exitCode,
				}
			}

			stage++
			pod.Status.EphemeralContainerStatuses[0].State = state

			return true, pod, nil
		})

		k := &Kubernetes{client: c}

		synctest.Test(t, func(t *testing.T) {
			code, err := k.PollEphemeralContainerStatus(t.Context(), pod, ephemeralContainer)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if code != exitCode {
				t.Errorf("expecting exit code %d, got %d", exitCode, code)
			}
		})
	})

	t.Run("ephemeral container not found", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      podName,
			},
			Status: corev1.PodStatus{
				EphemeralContainerStatuses: []corev1.ContainerStatus{},
			},
		}

		k := Kubernetes{
			client: testClient.NewClientset(pod),
		}

		synctest.Test(t, func(t *testing.T) {
			code, err := k.PollEphemeralContainerStatus(t.Context(), pod, ephemeralContainer)
			if !errors.Is(err, errEphemeralContainerNotFound) {
				t.Fatalf("expecting error %v, got %v", errEphemeralContainerNotFound, err)
			}
			if code != -1 {
				t.Errorf("expecting error code -1, got %d", code)
			}
		})
	})

	t.Run("timeout", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      podName,
			},
			Status: corev1.PodStatus{
				EphemeralContainerStatuses: []corev1.ContainerStatus{
					corev1.ContainerStatus{
						Name: ephemeralContainer,
						State: corev1.ContainerState{
							Running: &corev1.ContainerStateRunning{},
						},
					},
				},
			},
		}
		k := Kubernetes{
			client: testClient.NewClientset(pod),
		}

		synctest.Test(t, func(t *testing.T) {
			code, err := k.PollEphemeralContainerStatus(t.Context(), pod, ephemeralContainer)
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("expecting error %v, got %v", context.DeadlineExceeded, err)
			}
			if code != -1 {
				t.Errorf("expecting error code -1, got %d", code)
			}
		})
	})
}
