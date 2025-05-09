package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/grafana/nethax/pkg/common"
	"github.com/grafana/nethax/pkg/kubernetes"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ExecuteTest returns the execute-test command
func ExecuteTest() *cobra.Command {
	var testFile string

	cmd := &cobra.Command{
		Use:   "execute-test -f example/OtelDemoTestPlan.yaml",
		Short: "Execute network connectivity test plan",
		Run: func(cmd *cobra.Command, args []string) {
			if testFile == "" {
				cmd.Println("Error: test file must be specified")
				cmd.Help() //nolint:errcheck
				common.ExitConfigError()
			}

			file, err := os.Open(testFile)
			if err != nil {
				cmd.Printf("Error opening test file: %v\n", err)
				common.ExitConfigError()
			}
			defer file.Close() //nolint:errcheck

			plan, err := ParseTestPlan(file)
			if err != nil {
				cmd.Printf("Error parsing test plan: %v\n", err)
				common.ExitConfigError()
			}

			if !executeTest(cmd.Context(), plan) {
				common.ExitFailure()
			}
		},
	}

	cmd.Flags().StringVarP(&testFile, "file", "f", "", "Path to the test configuration YAML file")
	cmd.MarkFlagRequired("file") //nolint:errcheck

	return cmd
}

// indent prints the given string with the specified number of spaces for indentation
func indent(s string, level int) {
	fmt.Println(strings.Repeat("  ", level) + s)
}

func indentf(level int, format string, a ...any) {
	fmt.Print(strings.Repeat(" ", level))
	fmt.Printf(format, a...)
	fmt.Println()
}

func executeTest(ctx context.Context, plan *TestPlan) bool {
	indentf(0, "Test Plan: "+plan.Name)
	indentf(0, "Description: "+plan.Description)
	fmt.Println()

	k := kubernetes.GetKubernetes("")
	allTestsPassed := true

	for _, target := range plan.TestTargets {
		indentf(1, "Target: "+target.Name)
		indentf(1, "Selector: "+target.PodSelector)
		if target.Namespace != "" {
			indentf(1, "Namespace: "+target.Namespace)
		}
		indentf(1, "Selection Mode: "+target.PodSelection.Mode)

		// Find pods matching the selector
		pods, err := k.Client.CoreV1().Pods(target.Namespace).List(ctx, v1.ListOptions{
			LabelSelector: target.PodSelector,
		})
		if err != nil {
			indentf(1, "Error: Failed to find pods: %v", err)
			fmt.Println()
			allTestsPassed = false
			continue
		}

		if len(pods.Items) == 0 {
			indentf(1, "Error: No pods found matching selector %s", target.PodSelector)
			fmt.Println()
			allTestsPassed = false
			continue
		}

		// Select pods based on the selection mode
		var selectedPods []*corev1.Pod
		if mode := target.PodSelection.Mode; mode != "all" && mode != "random" {
			indentf(1, "Error: Invalid pod selection mode: %s", target.PodSelection.Mode)
			fmt.Println()
			allTestsPassed = false
			continue
		}

		// Only select pods that have Ready condition set to true
		for i := range pods.Items {
			if isPodReady(&pods.Items[i]) {
				selectedPods = append(selectedPods, &pods.Items[i])
			}
		}

		if len(selectedPods) == 0 {
			indentf(1, "Warning: No ready pods found matching selector %s", target.PodSelector)
			fmt.Println()
			allTestsPassed = false
			continue
		}

		if target.PodSelection.Mode == "random" {
			// Select one random pod from the ready pods
			randomIndex := rand.Intn(len(selectedPods))
			selectedPods = []*corev1.Pod{selectedPods[randomIndex]}
		}

		indentf(1, "Selected %d ready pod(s) for testing", len(selectedPods))

		// Execute tests for each selected pod
		for _, pod := range selectedPods {
			indentf(1, fmt.Sprintf("Pod: %s/%s", pod.Namespace, pod.Name))

			// Execute each test for this pod
			for _, test := range target.Tests {
				indentf(2, "Test: "+test.Name)
				indentf(3, "Endpoint: "+test.Endpoint)
				indentf(3, "Type: "+test.Type)
				indentf(3, "Expected Status: %d", test.StatusCode)
				indentf(3, "Expect Fail: %v", test.ExpectFail)
				indentf(3, "Timeout: %s", test.Timeout.String())

				// Parse the endpoint URL for HTTP tests
				if test.Type != "tcp" {
					_, err := url.Parse(test.Endpoint)
					if err != nil {
						indentf(3, "Error: Invalid endpoint URL: %v", err)
						fmt.Println()
						allTestsPassed = false
						continue
					}
				}

				// Prepare the test command
				command := []string{"/probe"}
				arguments := []string{
					"--url", test.Endpoint,
					"--timeout", test.Timeout.String(),
					"--expected-status", strconv.Itoa(test.StatusCode),
				}

				if test.Type == "tcp" {
					arguments = append(arguments, "--type", "tcp")
					if test.ExpectFail {
						arguments = append(arguments, "--expect-fail")
					}
				}

				// Launch ephemeral container to execute the test
				probedPod, probeContainerName, err := k.LaunchEphemeralContainer(ctx, pod, command, arguments)
				if err != nil {
					indentf(3, "Error: Failed to launch ephemeral probe container: %v", err)
					fmt.Println()
					allTestsPassed = false
					continue
				}

				// Wait for the test to complete and get the exit status
				exitStatus := k.PollEphemeralContainerStatus(ctx, probedPod, probeContainerName)

				// Check if the test passed based on the probe's exit status
				if exitStatus == 0 {
					indentf(3, "Result: PASSED")
					fmt.Println()
				} else {
					indentf(3, "Result: FAILED (exit code: %d)", exitStatus)
					fmt.Println()
					allTestsPassed = false
				}
			}
		}
	}

	return allTestsPassed
}

// isPodReady checks if a pod is ready by looking at its Ready condition
func isPodReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}
