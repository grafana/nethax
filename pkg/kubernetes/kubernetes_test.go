package kubernetes

import (
	"errors"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testClient "k8s.io/client-go/kubernetes/fake"
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
			ns, sel string
		}{
			{corev1.NamespaceAll, "app=redis"},
			{"nethax", "app=postgresql"},
			{"grafana", "app=nethax"},
			{"mimir", ""},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("ns=%s,sel=%s", tt.ns, tt.sel), func(t *testing.T) {
				pods, err := k.GetPods(t.Context(), tt.ns, tt.sel)
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
		pods, err := k.GetPods(t.Context(), "nethax", "app=nethax")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(pods) == 0 {
			t.Fatal("expecting pods, none found")
		}
	})
}

func TestLaunchEphemeralContainer(t *testing.T) {

	// prepare
	k := setup()
	name := "grafanyaa"
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name}}
	k.client.CoreV1().Namespaces().Create(t.Context(), namespace, metav1.CreateOptions{})
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
	actual, _, _ := k.LaunchEphemeralContainer(t.Context(), pod, []string{"nyaa"}, []string{"rawr"})

	// assert
	if len(actual.Spec.EphemeralContainers) != 1 {
		t.Fatalf("Expected PodSpec EphemeralContainers to be 1, got: %d", len(actual.Spec.EphemeralContainers))
	}

	// clean up
	k.client.CoreV1().Pods(name).Delete(t.Context(), name, metav1.DeleteOptions{})
	k.client.CoreV1().Namespaces().Delete(t.Context(), name, metav1.DeleteOptions{})
}

func TestGetEphemeralContainerExitStatus(t *testing.T) {
	const ephemeralContainerName = "nethaxme"

	tests := map[string]struct {
		statuses []corev1.ContainerStatus
		exp      int32
	}{
		"no ephemeral containers": {nil, -1},
		"ephemeral container not terminated": {
			[]corev1.ContainerStatus{
				{
					Name:  ephemeralContainerName,
					State: corev1.ContainerState{Terminated: nil},
				},
			},
			-1,
		},
		"other ephemeral container terminated": {
			[]corev1.ContainerStatus{
				{
					Name: "some-other-c",
					State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							ExitCode: 0,
						},
					},
				},
			},
			-1,
		},
		"ephemeral container terminated": {
			[]corev1.ContainerStatus{
				{
					Name: "some-other-c",
					State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							ExitCode: 0,
						},
					},
				},
				{
					Name: ephemeralContainerName,
					State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							ExitCode: 2,
						},
					},
				},
			},
			2,
		},
	}

	for n, tt := range tests {
		t.Run(n, func(t *testing.T) {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "mimir-dev-013",
					Name:      "ingester",
				},
				Status: corev1.PodStatus{
					EphemeralContainerStatuses: tt.statuses,
				},
			}

			k := &Kubernetes{
				client: testClient.NewSimpleClientset(pod),
			}

			got, err := k.getEphemeralContainerExitStatus(t.Context(), pod, ephemeralContainerName)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.exp != got {
				t.Errorf("expecting exist status %d, got %d", tt.exp, got)
			}
		})
	}

	t.Run("error", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "mimir-dev-013",
				Name:      "ingester",
			},
		}

		k := &Kubernetes{
			client: testClient.NewSimpleClientset(),
		}

		got, err := k.getEphemeralContainerExitStatus(t.Context(), pod, ephemeralContainerName)
		if err == nil {
			t.Fatal("expecting error, got nil")
		}
		if got != -1 {
			t.Fatalf("expecting code -1, got %d", got)
		}
	})
}

func TestIsEphemeralContainerTerminated(t *testing.T) {
	const ephemeralContainerName = "nethaxme"

	tests := map[string]struct {
		statuses []corev1.ContainerStatus
		exp      bool
	}{
		"no ephemeral containers": {nil, false},
		"ephemeral container not terminated": {
			[]corev1.ContainerStatus{
				{
					Name:  ephemeralContainerName,
					State: corev1.ContainerState{Terminated: nil},
				},
			},
			false,
		},
		"other ephemeral container terminated": {
			[]corev1.ContainerStatus{
				{
					Name: "some-other-c",
					State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							ExitCode: 0,
						},
					},
				},
			},
			false,
		},
		"ephemeral container terminated": {
			[]corev1.ContainerStatus{
				{
					Name: "some-other-c",
					State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							ExitCode: 0,
						},
					},
				},
				{
					Name: ephemeralContainerName,
					State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							ExitCode: 2,
						},
					},
				},
			},
			true,
		},
	}

	for n, tt := range tests {
		t.Run(n, func(t *testing.T) {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "mimir-dev-013",
					Name:      "ingester",
				},
				Status: corev1.PodStatus{
					EphemeralContainerStatuses: tt.statuses,
				},
			}

			k := &Kubernetes{
				client: testClient.NewSimpleClientset(pod),
			}

			ok, err := k.isEphemeralContainerTerminated(pod, ephemeralContainerName)(t.Context())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.exp != ok {
				t.Errorf("expecting container terminated %t, got %t", tt.exp, ok)
			}
		})
	}

	t.Run("errors", func(t *testing.T) {
		t.Run("pod not found", func(t *testing.T) {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "mimir-dev-013",
					Name:      "ingester",
				},
			}

			k := &Kubernetes{
				client: testClient.NewSimpleClientset(),
			}

			ok, err := k.isEphemeralContainerTerminated(pod, ephemeralContainerName)(t.Context())
			if err == nil {
				t.Fatal("expecting error, got nil")
			}
			if ok {
				t.Errorf("expecting container terminated, got %t", ok)
			}
		})

		for n, s := range map[string]corev1.ContainerState{
			"waiting": {
				Waiting: &corev1.ContainerStateWaiting{},
			},
			"running": {
				Running: &corev1.ContainerStateRunning{},
			},
		} {
			t.Run(n, func(t *testing.T) {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "mimir-dev-013",
						Name:      "ingester",
					},
					Status: corev1.PodStatus{
						EphemeralContainerStatuses: []corev1.ContainerStatus{
							{
								Name:  ephemeralContainerName,
								State: s,
							},
						},
					},
				}

				k := &Kubernetes{
					client: testClient.NewSimpleClientset(pod),
				}

				ok, err := k.isEphemeralContainerTerminated(pod, ephemeralContainerName)(t.Context())
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if ok {
					t.Errorf("expecting container terminated to be false")
				}
			})
		}
	})
}

func TestPollEphemeralContainerStatus(t *testing.T) {
	t.Run("container terminated", func(t *testing.T) {
		const ephemeralContainerName = "nethaxme"

		for _, exitCode := range []int32{0, 2} {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "mimir-dev-013",
					Name:      "ingester",
				},
				Status: corev1.PodStatus{
					EphemeralContainerStatuses: []corev1.ContainerStatus{
						{
							Name: ephemeralContainerName,
							State: corev1.ContainerState{
								Terminated: &corev1.ContainerStateTerminated{
									ExitCode: exitCode,
								},
							},
						},
					},
				},
			}

			k := &Kubernetes{
				client: testClient.NewSimpleClientset(pod),
			}

			got, err := k.PollEphemeralContainerStatus(t.Context(), pod, ephemeralContainerName)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if exitCode != got {
				t.Fatalf("expecting container status %d, got %d", exitCode, got)
			}
		}
	})

	t.Run("container not terminated", func(t *testing.T) {
		states := map[string]corev1.ContainerState{
			"running": {
				Running: &corev1.ContainerStateRunning{
					StartedAt: metav1.Now(),
				},
			},
			"waiting": {
				Waiting: &corev1.ContainerStateWaiting{},
			},
		}

		for n, state := range states {
			t.Run("state="+n, func(t *testing.T) {
				const ephemeralContainerName = "nethaxme"

				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "mimir-dev-013",
						Name:      "ingester",
					},
					Status: corev1.PodStatus{
						EphemeralContainerStatuses: []corev1.ContainerStatus{
							{
								Name:  ephemeralContainerName,
								State: state,
							},
						},
					},
				}

				k := &Kubernetes{
					client: testClient.NewSimpleClientset(pod),
				}

				code, err := k.PollEphemeralContainerStatus(t.Context(), pod, ephemeralContainerName)
				if err == nil {
					t.Fatal("expecting error")
				}
				if code != -1 {
					t.Fatalf("expecting -1 exit code, got %d", code)
				}
			})
		}
	})
}
