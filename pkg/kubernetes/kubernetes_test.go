package kubernetes

import (
	"context"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testClient "k8s.io/client-go/kubernetes/fake"
)

func setup() *Kubernetes {
	k := GetKubernetes()
	k.Client = testClient.NewSimpleClientset()
	return k
}

func TestGetPods(t *testing.T) {
	// prepare
	k := setup()
	expected := []string{"grafanyaa"}
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: expected[0]}}
	k.Client.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: expected[0],
			Name:      expected[0],
		},
	}
	k.Client.CoreV1().Pods(expected[0]).Create(context.TODO(), pod, metav1.CreateOptions{})

	// execute
	actual := GetPods(expected[0])

	// assert
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Wanted %s, got %s", expected, actual)
	}

	// clean up
	k.Client.CoreV1().Pods(expected[0]).Delete(context.TODO(), expected[0], metav1.DeleteOptions{})
	k.Client.CoreV1().Namespaces().Delete(context.TODO(), expected[0], metav1.DeleteOptions{})
}

func TestLaunchEphemeralContainer(t *testing.T) {
	// prepare
	k := setup()
	name := "grafanyaa"
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name}}
	k.Client.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
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
	k.Client.CoreV1().Pods(name).Create(context.TODO(), pod, metav1.CreateOptions{})

	// execute
	actual, _, _ := LaunchEphemeralContainer(pod, []string{"nyaa"}, []string{"rawr"})

	// assert
	if len(actual.Spec.EphemeralContainers) != 1 {
		t.Fatalf("Expected PodSpec EphemeralContainers to be 1, got: %d", len(actual.Spec.EphemeralContainers))
	}

	// clean up
	k.Client.CoreV1().Pods(name).Delete(context.TODO(), name, metav1.DeleteOptions{})
	k.Client.CoreV1().Namespaces().Delete(context.TODO(), name, metav1.DeleteOptions{})
}
