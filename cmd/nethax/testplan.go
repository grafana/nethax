package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

// Test represents a single network connectivity test
type Test struct {
	Name       string        `yaml:"name"`
	Endpoint   string        `yaml:"endpoint"`
	StatusCode int           `yaml:"statusCode"`
	Type       TestType      `yaml:"type,omitempty"`
	ExpectFail bool          `yaml:"expectFail,omitempty"`
	Timeout    time.Duration `yaml:"timeout"`
	ProbeImage string        `yaml:"probeImage,omitempty"`
}

// PodSelector represents how pods should be selected for testing
type PodSelector struct {
	Mode   SelectionMode `yaml:"mode"` // "all" or "random"
	Labels string        `yaml:"labels"`
	Fields string        `yaml:"fields"`
}

func (s PodSelector) String() string {
	var b strings.Builder
	b.WriteString("mode: ")
	b.WriteString(string(s.Mode))

	if s.Labels != "" {
		b.WriteString(", labels: ")
		b.WriteString(s.Labels)
	}
	if s.Fields != "" {
		b.WriteString(", fields: ")
		b.WriteString(s.Fields)
	}

	return b.String()
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
	if err := yaml.NewDecoder(reader, yaml.Strict()).Decode(&plan); err != nil {
		return nil, err
	}

	return &plan.TestPlan, nil
}

type TestType int

func (tt TestType) String() string {
	switch tt {
	case TestTypeHTTP:
		return "http"
	case TestTypeTCP:
		return "tcp"
	case TestTypeDNS:
		return "dns"
	default:
		panic("unrecognized testype")
	}
}

const (
	TestTypeHTTP TestType = iota
	TestTypeTCP
	TestTypeDNS
)

var errInvalidTestType = errors.New("invalid test type")

func yamlUnmarshalTestType(tt *TestType, b []byte) error {
	switch strings.TrimSpace(strings.ToLower(string(b))) {
	case "", "http", "https":
		*tt = TestTypeHTTP
	case "tcp":
		*tt = TestTypeTCP
	case "dns":
		*tt = TestTypeDNS
	default:
		return fmt.Errorf("%w: %q", errInvalidTestType, b)
	}

	return nil
}

func init() {
	yaml.RegisterCustomUnmarshaler(yamlUnmarshalTestType)
	yaml.RegisterCustomUnmarshaler(yamlUnmarshalSelectionMode)
}

type SelectionMode string

const (
	SelectionModeAll    SelectionMode = "all"
	SelectionModeRandom SelectionMode = "random"
)

func yamlUnmarshalSelectionMode(m *SelectionMode, b []byte) error {
	// we need to add the cases with quotes because YAML is such a
	// good language that `foo: bar` is the same as `foo: "bar"` and
	// also the same as `foo: 'bar'` but, hey, the three values are
	// passed as-is to the parser and thus have to take into account
	// these characters.
	switch strings.TrimSpace(strings.ToLower(string(b))) {
	case "all", "\"all\"", "'all'":
		*m = SelectionModeAll
	case "random", "\"random\"", "'random'":
		*m = SelectionModeRandom
	default:
		return fmt.Errorf("%w: %q", errInvalidSelectionMode, b)
	}

	return nil
}
