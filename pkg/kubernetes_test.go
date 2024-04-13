package pkg

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
func TestGetNamespaces(t *testing.T) {
	k := setup()

	expected := []string{"default"}
	namespaceName := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: expected[0]}}
	k.Client.CoreV1().Namespaces().Create(context.TODO(), namespaceName, metav1.CreateOptions{})
	actual := GetNamespaces()
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Wanted %s, got %s", expected, actual)
	}

}
func teardown() {}
