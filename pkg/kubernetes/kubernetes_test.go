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
