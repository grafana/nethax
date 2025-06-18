package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/grafana/nethax/pkg/kubernetes"
	pf "github.com/grafana/nethax/pkg/probeflags"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

// ExecuteTest returns the execute-test command
func ExecuteTest() *cobra.Command {
	var testFile, defaultProbeImage, kontext string

	cmd := &cobra.Command{
		Use:   "execute-test -f example/OtelDemoTestPlan.yaml",
		Short: "Execute network connectivity test plan",
		Run: func(cmd *cobra.Command, args []string) {
			if testFile == "" {
				cmd.Println("Error: test file must be specified")
				cmd.Help() //nolint:errcheck
				os.Exit(exitCodeConfigError)
			}

			file, err := os.Open(testFile)
			if err != nil {
				cmd.Printf("Error opening test file: %v\n", err)
				os.Exit(exitCodeConfigError)
			}
			defer file.Close() //nolint:errcheck

			plan, err := ParseTestPlan(file)
			if err != nil {
				cmd.Printf("Error parsing test plan: %v\n", err)
				os.Exit(exitCodeConfigError)
			}

			k, err := kubernetes.New(kontext)
			if err != nil {
				cmd.Printf("Error creating Kubernetes client: %v\n", err)
				os.Exit(exitCodeConfigError)
			}

			kubernetes.DefaultProbeImage = defaultProbeImage
			if !executeTest(cmd.Context(), k, plan) {
				os.Exit(exitCodeFailure)
			}
		},
	}

	cmd.Flags().StringVarP(&testFile, "file", "f", "", "Path to the test configuration YAML file")
	cmd.MarkFlagRequired("file") //nolint:errcheck

	cmd.Flags().StringVarP(&kontext, "context", "c", "", "Kubernetes context to connect. Leave empty for in-cluster context.")

	cmd.Flags().StringVar(&defaultProbeImage,
		"default-probe-image",
		kubernetes.DefaultProbeImage,
		"Default probe image to use if test plan doesn't specify one.",
	)

	return cmd
}

func indent(level int, format string, a ...any) {
	fmt.Print(strings.Repeat(" ", level))
	fmt.Printf(format, a...)
	fmt.Println()
}

func executeTest(ctx context.Context, k *kubernetes.Kubernetes, plan *TestPlan) bool {
	indent(0, "Test Plan: %s", plan.Name)
	indent(0, "Description: %s", plan.Description)
	fmt.Println()

	allTestsPassed := true

	for _, target := range plan.TestTargets {
		indent(1, "Target: %s", target.Name)
		indent(1, "Selector: %s", target.PodSelector)
		if target.Namespace != "" {
			indent(1, "Namespace: %s", target.Namespace)
		}

		selectedPods, err := findPods(ctx, k, target.Namespace, target.PodSelector)
		if err != nil {
			indent(1, "Error: %v", err)
			fmt.Println()
			allTestsPassed = false
			continue
		}

		indent(1, "Selected %d ready pod(s) for testing", len(selectedPods))

		// Execute tests for each selected pod
		for _, pod := range selectedPods {
			indent(1, "Pod: %s/%s", pod.Namespace, pod.Name)

			// Execute each test for this pod
			for _, test := range target.Tests {
				indent(2, "Test: %s", test.Name)
				indent(3, "Endpoint: %s", test.Endpoint)
				indent(3, "Type: %s", test.Type)
				indent(3, "Expected Status: %d", test.StatusCode)
				indent(3, "Expect Fail: %v", test.ExpectFail)
				indent(3, "Timeout: %s", test.Timeout.String())

				// Parse the endpoint URL for HTTP tests
				if test.Type != TestTypeTCP {
					_, err := url.Parse(test.Endpoint)
					if err != nil {
						indent(3, "Error: Invalid endpoint URL: %v", err)
						fmt.Println()
						allTestsPassed = false
						continue
					}
				}

				// Prepare the test command
				command := []string{"/nethax-probe"}
				arguments := []string{
					pf.Flagify(pf.ArgURL), test.Endpoint,
					pf.Flagify(pf.ArgTimeout), test.Timeout.String(),
					pf.Flagify(pf.ArgExpectedStatus), strconv.Itoa(test.StatusCode),
				}

				if test.Type == TestTypeTCP {
					arguments = append(arguments, pf.Flagify(pf.ArgType), pf.TestTypeTCP)
					if test.ExpectFail {
						arguments = append(arguments, pf.Flagify(pf.ArgExpectFail))
					}
				}

				indent(3, "Probe Image: '%s'", kubernetes.GetProbeImage(test.ProbeImage))

				// Launch ephemeral container to execute the test
				probedPod, probeContainerName, err := k.LaunchEphemeralContainer(ctx, &pod, test.ProbeImage, command, arguments)
				if err != nil {
					indent(3, "Error: Failed to launch ephemeral probe container: %v", err)
					fmt.Println()
					allTestsPassed = false
					continue
				}

				// Wait for the test to complete and get the exit status
				exitStatus, err := k.PollEphemeralContainerStatus(ctx, probedPod, probeContainerName)
				if err != nil {
					indent(3, "Result: ERROR %v", err)
					fmt.Println()
					allTestsPassed = false
					continue
				}

				// Check if the test passed based on the probe's exit status
				if exitStatus == 0 {
					indent(3, "Result: PASSED")
					fmt.Println()
				} else {
					indent(3, "Result: FAILED (exit code: %d)", exitStatus)
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

func findPods(ctx context.Context, k *kubernetes.Kubernetes, namespace string, selector PodSelector) ([]corev1.Pod, error) {
	pods, err := k.GetPods(ctx, namespace, selector.Labels, selector.Fields)
	if err != nil {
		return nil, fmt.Errorf("failed to find pods: %w", err)
	}

	return selectPods(selector.Mode, pods)
}

var (
	errInvalidSelectionMode = errors.New("invalid pod selection mode")
	errNoReadyPods          = errors.New("no ready pods found")
)

func selectPods(mode SelectionMode, pods []corev1.Pod) ([]corev1.Pod, error) {
	if mode != SelectionModeAll && mode != SelectionModeRandom {
		return nil, fmt.Errorf("%w: %s", errInvalidSelectionMode, mode)
	}

	var selectedPods []corev1.Pod

	// Only select pods that have Ready condition set to true
	for i := range pods {
		if isPodReady(&pods[i]) {
			selectedPods = append(selectedPods, pods[i])
		}
	}

	if len(selectedPods) == 0 {
		return nil, errNoReadyPods
	}

	// Select pods based on the selection mode
	if mode == SelectionModeRandom {
		// Select one random pod from the ready pods
		randomIndex := rand.Intn(len(selectedPods))
		selectedPods = []corev1.Pod{selectedPods[randomIndex]}
	}

	return selectedPods, nil
}
