package main

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"testing"
)

//go:embed testdata/example.yml
var exampleTestPlan string

func TestParseTestPlan(t *testing.T) {
	r := strings.NewReader(exampleTestPlan)

	tp, err := ParseTestPlan(r)
	if err != nil {
		t.Fatal(err)
	}

	if e, g := 2, len(tp.TestTargets); e != g {
		t.Fatalf("expecting %d test targets, got %d", e, g)
	}

	t.Run("default to HTTP(S)", func(t *testing.T) {
		tt := tp.TestTargets[0]

		for i, tt := range tt.Tests {
			if tt.Type != TestTypeHTTP {
				t.Errorf("expecting test target 0, test %d to be of type HTTP(S), got %q", i, tt.Type)
			}
		}
	})

	t.Run("don't override type", func(t *testing.T) {
		tt := tp.TestTargets[1].Tests[1]

		if e, g := TestTypeTCP, tt.Type; e != g {
			t.Fatalf("expecting type to be %q, got %q", e, g)
		}
	})

	t.Run("parse probe image", func(t *testing.T) {
		// Check that most tests don't have a probe image
		tt := tp.TestTargets[0].Tests[0]
		if tt.ProbeImage != "" {
			t.Errorf("expecting first test to have empty probe image, got %q", tt.ProbeImage)
		}

		// Check that the test with custom probe image is parsed correctly
		tt = tp.TestTargets[0].Tests[2]
		if e, g := "myregistry.io/custom-probe:v2.0.0", tt.ProbeImage; e != g {
			t.Errorf("expecting probe image to be %q, got %q", e, g)
		}
	})
}

func TestPodSelector_String(t *testing.T) {
	tests := []struct {
		sel PodSelector
		exp string
	}{
		{PodSelector{Mode: "all"}, "mode: all"},
		{PodSelector{Mode: "random"}, "mode: random"},

		{PodSelector{Mode: "all", Labels: "app=nethax"}, "mode: all, labels: app=nethax"},
		{PodSelector{Mode: "all", Fields: "spec.nodeName=foo-bar-42"}, "mode: all, fields: spec.nodeName=foo-bar-42"},
		{PodSelector{Mode: "all", Labels: "app=nethax", Fields: "spec.nodeName=foo-bar-42"}, "mode: all, labels: app=nethax, fields: spec.nodeName=foo-bar-42"},

		{PodSelector{Mode: "random", Labels: "app=grafana"}, "mode: random, labels: app=grafana"},
		{PodSelector{Mode: "random", Fields: "spec.nodeName=foo-bar-23"}, "mode: random, fields: spec.nodeName=foo-bar-23"},
		{PodSelector{Mode: "random", Labels: "app=grafana", Fields: "spec.nodeName=foo-bar-23"}, "mode: random, labels: app=grafana, fields: spec.nodeName=foo-bar-23"},
	}

	for _, tt := range tests {
		// using fmt.Sprint here so we know it uses Stringer
		if got := fmt.Sprint(tt.sel); tt.exp != got {
			t.Errorf("for %#v expecting %q, got %q", tt.sel, tt.exp, got)
		}
	}
}

func TestTestType(t *testing.T) {
	// We default to HTTP
	var tt TestType

	if tt != TestTypeHTTP {
		t.Fatalf("unexpected zero value: %d", tt)
	}
}

func TestTestType_UnmarshalYAML(t *testing.T) {
	tests := map[string]struct {
		exp TestType
		err error
	}{
		"":      {TestTypeHTTP, nil},
		"http":  {TestTypeHTTP, nil},
		"https": {TestTypeHTTP, nil},
		"tcp":   {TestTypeTCP, nil},
		"dns":   {TestTypeDNS, nil},
		// ignore case
		"HTTP":  {TestTypeHTTP, nil},
		"HTTPS": {TestTypeHTTP, nil},
		"TCP":   {TestTypeTCP, nil},
		"DNS":   {TestTypeDNS, nil},
		// invalid values // TODO(inkel) this could probably be a fuzz test
		"foo": {TestTypeHTTP, errInvalidTestType},
	}

	for in, tt := range tests {
		t.Run("in="+in, func(t *testing.T) {
			var got TestType

			err := yamlUnmarshalTestType(&got, []byte(in))
			if !errors.Is(err, tt.err) {
				t.Fatalf("expecting error %v, got %v", tt.err, err)
			}
			if tt.exp != got {
				t.Fatalf("expecting TestType %d (%[1]s) got %d (%[2]s)", tt.exp, got)
			}
		})
	}
}

func TestSelectionMode_UnmarshalYAML(t *testing.T) {
	tests := map[string]struct {
		exp SelectionMode
		err error
	}{
		// cannot be empty
		"": {"", errInvalidSelectionMode},

		"all": {SelectionModeAll, nil},
		"ALL": {SelectionModeAll, nil},

		"random": {SelectionModeRandom, nil},
		"rAnDom": {SelectionModeRandom, nil},
	}

	for in, tt := range tests {
		t.Run("in="+in, func(t *testing.T) {
			var got SelectionMode

			err := yamlUnmarshalSelectionMode(&got, []byte(in))
			if !errors.Is(err, tt.err) {
				t.Fatalf("expecting error %v, got %v", tt.err, err)
			}
			if tt.exp != got {
				t.Fatalf("expecting SelectionMode %q, got %q", tt.exp, got)
			}
		})
	}
}
