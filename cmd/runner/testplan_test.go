package main

import (
	_ "embed"
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
			if tt.Type != "HTTP(S)" {
				t.Errorf("expecting test target 0, test %d to be of type HTTP(S), got %q", i, tt.Type)
			}
		}
	})

	t.Run("don't override type", func(t *testing.T) {
		tt := tp.TestTargets[1].Tests[1]

		if e, g := "tcp", tt.Type; e != g {
			t.Fatalf("expecting type to be %q, got %q", e, g)
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
