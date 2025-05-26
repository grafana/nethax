package main

import (
	_ "embed"
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
