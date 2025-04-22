package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/grafana/nethax/pkg/kubernetes"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Test represents a single network connectivity test
type Test struct {
	Name       string        `yaml:"name"`
	Endpoint   string        `yaml:"endpoint"`
	StatusCode int           `yaml:"statusCode"`
	Type       string        `yaml:"type,omitempty"`
	ExpectFail bool          `yaml:"expectFail,omitempty"`
	Timeout    time.Duration `yaml:"timeout"`
}

// PodSelection represents how pods should be selected for testing
type PodSelection struct {
	Mode string `yaml:"mode"` // "all" or "random"
}

// TestTarget represents a pod target with multiple tests
type TestTarget struct {
	Name         string       `yaml:"name"`
	PodSelector  string       `yaml:"podSelector"`
	Namespace    string       `yaml:"namespace,omitempty"`
	PodSelection PodSelection `yaml:"podSelection"`
	Tests        []Test       `yaml:"tests"`
}

// TestPlan represents a collection of test targets with metadata
type TestPlan struct {
	Name        string       `yaml:"name"`
	Description string       `yaml:"description"`
	TestTargets []TestTarget `yaml:"testTargets"`
}

// ParseTestPlan reads YAML content and returns a TestPlan
func ParseTestPlan(reader io.Reader) (*TestPlan, error) {
	var plan struct {
		TestPlan TestPlan `yaml:"testPlan"`
	}
	if err := yaml.NewDecoder(reader).Decode(&plan); err != nil {
		return nil, err
	}

	// Set default test type to HTTP(S)
	for i, target := range plan.TestPlan.TestTargets {
		for j, test := range target.Tests {
			if test.Type == "" {
				plan.TestPlan.TestTargets[i].Tests[j].Type = "HTTP(S)"
			}
		}
	}
	return &plan.TestPlan, nil
}

// GetTimeoutDuration converts the timeout in seconds to a time.Duration
func (t *Test) GetTimeoutDuration() time.Duration {
	return t.Timeout
}

// ExecuteTest returns the execute-test command
func ExecuteTest() *cobra.Command {
	var testFile string

	cmd := &cobra.Command{
		Use:   "execute-test -f example/OtelDemoTestPlan.yaml",
		Short: "Execute network connectivity test plan",
		Run: func(cmd *cobra.Command, args []string) {
			if testFile == "" {
				cmd.Println("Error: test file must be specified")
				cmd.Help()
				os.Exit(2)
			}

			file, err := os.Open(testFile)
			if err != nil {
				cmd.Printf("Error opening test file: %v\n", err)
				os.Exit(2)
			}
			defer file.Close()

			plan, err := ParseTestPlan(file)
			if err != nil {
				cmd.Printf("Error parsing test plan: %v\n", err)
				os.Exit(2)
			}

			if !executeTest(cmd.Context(), plan) {
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&testFile, "file", "f", "", "Path to the test configuration YAML file")
	cmd.MarkFlagRequired("file")

	return cmd
}

// indent prints the given string with the specified number of spaces for indentation
func indent(s string, level int) {
	fmt.Println(strings.Repeat("  ", level) + s)
}

func executeTest(ctx context.Context, plan *TestPlan) bool {
	indent("Test Plan: "+plan.Name, 0)
	indent("Description: "+plan.Description, 0)
	fmt.Println()

	k := kubernetes.GetKubernetes()
	allTestsPassed := true

	for _, target := range plan.TestTargets {
		indent("Target: "+target.Name, 1)
		indent("Selector: "+target.PodSelector, 1)
		if target.Namespace != "" {
			indent("Namespace: "+target.Namespace, 1)
		}
		indent("Selection Mode: "+target.PodSelection.Mode, 1)

		// Find pods matching the selector
		pods, err := k.Client.CoreV1().Pods(target.Namespace).List(ctx, v1.ListOptions{
			LabelSelector: target.PodSelector,
		})
		if err != nil {
			indent(fmt.Sprintf("Error: Failed to find pods: %v", err), 1)
			fmt.Println()
			allTestsPassed = false
			continue
		}

		if len(pods.Items) == 0 {
			indent(fmt.Sprintf("Error: No pods found matching selector %s", target.PodSelector), 1)
			fmt.Println()
			allTestsPassed = false
			continue
		}

		// Select pods based on the selection mode
		var selectedPods []*corev1.Pod
		if mode := target.PodSelection.Mode; mode != "all" && mode != "random" {
			indent(fmt.Sprintf("Error: Invalid pod selection mode: %s", target.PodSelection.Mode), 1)
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
			indent(fmt.Sprintf("Warning: No ready pods found matching selector %s", target.PodSelector), 1)
			fmt.Println()
			allTestsPassed = false
			continue
		}

		if target.PodSelection.Mode == "random" {
			// Select one random pod from the ready pods
			randomIndex := rand.Intn(len(selectedPods))
			selectedPods = []*corev1.Pod{selectedPods[randomIndex]}
		}

		indent(fmt.Sprintf("Selected %d ready pod(s) for testing", len(selectedPods)), 1)

		// Execute tests for each selected pod
		for _, pod := range selectedPods {
			indent(fmt.Sprintf("Pod: %s/%s", pod.Namespace, pod.Name), 1)

			// Execute each test for this pod
			for _, test := range target.Tests {
				indent("Test: "+test.Name, 2)
				indent("Endpoint: "+test.Endpoint, 3)
				indent("Type: "+test.Type, 3)
				indent(fmt.Sprintf("Expected Status: %d", test.StatusCode), 3)
				indent(fmt.Sprintf("Expect Fail: %v", test.ExpectFail), 3)
				indent(fmt.Sprintf("Timeout: %s", test.Timeout.String()), 3)

				// Parse the endpoint URL for HTTP tests
				if test.Type != "tcp" {
					_, err := url.Parse(test.Endpoint)
					if err != nil {
						indent(fmt.Sprintf("Error: Invalid endpoint URL: %v", err), 3)
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
				probedPod, probeContainerName, err := kubernetes.LaunchEphemeralContainer(ctx, pod, command, arguments)
				if err != nil {
					indent(fmt.Sprintf("Error: Failed to launch ephemeral probe container: %v", err), 3)
					fmt.Println()
					allTestsPassed = false
					continue
				}

				// Wait for the test to complete and get the exit status
				exitStatus := kubernetes.PollEphemeralContainerStatus(ctx, probedPod, probeContainerName)

				// Check if the test passed based on the probe's exit status
				if exitStatus == 0 {
					indent("Result: PASSED", 3)
					fmt.Println()
				} else {
					indent(fmt.Sprintf("Result: FAILED (exit code: %d)", exitStatus), 3)
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
