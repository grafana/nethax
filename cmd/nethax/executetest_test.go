package main

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsPodReady(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		tests := [][]corev1.PodCondition{
			{
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
			},
			{
				{Type: corev1.ContainersReady, Status: corev1.ConditionFalse},
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
			},
			{
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
				{Type: corev1.ContainersReady, Status: corev1.ConditionFalse},
			},
			{
				{Type: corev1.PodInitialized, Status: corev1.ConditionUnknown},
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
			},
		}

		for _, conds := range tests {
			pod := &corev1.Pod{
				Status: corev1.PodStatus{
					Conditions: conds,
				},
			}
			if !isPodReady(pod) {
				t.Error("pod should be ready")
				for _, c := range conds {
					t.Logf("\tType: %s, Status: %s", c.Type, c.Status)
				}
			}
		}
	})

	t.Run("false", func(t *testing.T) {
		tests := [][]corev1.PodCondition{
			{
				{Type: corev1.PodReady, Status: corev1.ConditionFalse},
			},
			{
				{Type: corev1.PodReady, Status: corev1.ConditionUnknown},
			},
			{
				{Type: corev1.ContainersReady, Status: corev1.ConditionTrue},
				{Type: corev1.PodReady, Status: corev1.ConditionFalse},
			},
			{
				{Type: corev1.PodReady, Status: corev1.ConditionFalse},
				{Type: corev1.ContainersReady, Status: corev1.ConditionTrue},
			},
			{
				{Type: corev1.PodInitialized, Status: corev1.ConditionTrue},
				{Type: corev1.PodReady, Status: corev1.ConditionFalse},
			},
			{
				{Type: corev1.ContainersReady, Status: corev1.ConditionTrue},
			},
			{
				{Type: corev1.PodInitialized, Status: corev1.ConditionTrue},
			},
		}

		for _, conds := range tests {
			pod := &corev1.Pod{
				Status: corev1.PodStatus{
					Conditions: conds,
				},
			}
			if isPodReady(pod) {
				t.Error("pod should not be ready")
				for _, c := range conds {
					t.Logf("\tType: %s, Status: %s", c.Type, c.Status)
				}
			}
		}
	})
}

func mockPods(conds []corev1.PodCondition) []corev1.Pod {
	pods := make([]corev1.Pod, len(conds))
	for i, c := range conds {
		pods[i] = corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("pod-%03d", i),
			},
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{c},
			},
		}
	}
	return pods
}

func TestSelectPods(t *testing.T) {
	// inputs
	unreadyPods := mockPods([]corev1.PodCondition{
		{Type: corev1.PodReady, Status: corev1.ConditionFalse},
	})
	allPodsReady := mockPods([]corev1.PodCondition{
		{Type: corev1.PodReady, Status: corev1.ConditionTrue},
		{Type: corev1.PodReady, Status: corev1.ConditionTrue},
		{Type: corev1.PodReady, Status: corev1.ConditionTrue},
		{Type: corev1.PodReady, Status: corev1.ConditionTrue},
	})
	somePodsReady := mockPods([]corev1.PodCondition{
		{Type: corev1.PodReady, Status: corev1.ConditionTrue},
		{Type: corev1.PodReady, Status: corev1.ConditionFalse},
		{Type: corev1.PodReady, Status: corev1.ConditionFalse},
		{Type: corev1.PodReady, Status: corev1.ConditionTrue},
	})

	// naive pod comparison for checks
	cmp := func(a, b corev1.Pod) bool {
		return a.Name == b.Name
	}

	tests := map[string]struct {
		mode SelectionMode
		pods []corev1.Pod
		exp  []corev1.Pod
		err  error
	}{
		// errors
		"invalid mode":  {"foo", nil, nil, errInvalidSelectionMode},
		"no ready pods": {"all", unreadyPods, nil, errNoReadyPods},

		// all pods selection mode
		"all pods": {"all", allPodsReady, allPodsReady, nil},
		"all ready pods": {"all", somePodsReady, []corev1.Pod{
			somePodsReady[0], somePodsReady[3],
		}, nil},

		// random pod selection mode, see notes in switch below
		"random pod from all pods":  {"random", allPodsReady, nil, nil},
		"random pod from some pods": {"random", somePodsReady, nil, nil},
	}

	for n, tt := range tests {
		t.Run(n, func(t *testing.T) {
			got, err := selectPods(tt.mode, tt.pods)
			if !errors.Is(err, tt.err) {
				t.Fatalf("expecting error %v, got %v", tt.err, err)
			}

			// only ready pods selected
			for i, p := range got {
				if !isPodReady(&p) {
					t.Fatalf("found non-ready pod %v at index %d", p, i)
				}
			}

			switch tt.mode {
			case "random":
				// special case, we want only 1 pod returned, and that
				// pod should exist in the input pods.
				if len(got) != 1 {
					t.Fatalf("expecting 1 pod returned, got %d -> %v", len(got), got)
				}
				if exp := got[0]; !slices.ContainsFunc(got, func(pod corev1.Pod) bool {
					return cmp(exp, pod)
				}) {
					t.Fatalf("returned pod not found in input\nin: %v\ngot: %v", tt.pods, exp)
				}

			case "all":
				eq := slices.EqualFunc(tt.exp, got, func(a, b corev1.Pod) bool {
					return cmp(a, b)
				})
				if !eq {
					t.Fatalf("unexpected or missing pods\nexp: %v\ngot: %v", dumpPods(tt.exp), dumpPods(got))
				}
			}
		})

	}
}

// helpers for nicer dumping of pods
type podCond struct {
	Type   corev1.PodConditionType
	Status corev1.ConditionStatus
}

type podDump struct {
	Name  string
	Conds []podCond
}

func (d podDump) String() string { return fmt.Sprintf("{Name: %s, Conds: %v}", d.Name, d.Conds) }

func dumpPods(pods []corev1.Pod) []podDump {
	res := make([]podDump, len(pods))
	for i, p := range pods {
		d := podDump{Name: p.Name}
		for _, s := range p.Status.Conditions {
			d.Conds = append(d.Conds, podCond{s.Type, s.Status})
		}
		res[i] = d
	}
	return res
}
