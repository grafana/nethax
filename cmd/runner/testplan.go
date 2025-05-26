package main

import (
	"io"
	"time"

	"github.com/goccy/go-yaml"
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

// PodSelector represents how pods should be selected for testing
type PodSelector struct {
	Mode   string `yaml:"mode"` // "all" or "random"
	Labels string `yaml:"labels"`
}

// TestTarget represents a pod target with multiple tests
type TestTarget struct {
	Name        string      `yaml:"name"`
	Namespace   string      `yaml:"namespace,omitempty"`
	PodSelector PodSelector `yaml:"podSelector"`
	Tests       []Test      `yaml:"tests"`
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
