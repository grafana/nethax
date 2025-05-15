package main

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
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
