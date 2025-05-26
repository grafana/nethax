package probeflags

import (
	"flag"
	"os"
	"testing"
)

func TestFlagify(t *testing.T) {
	testArg := "test"
	var test bool
	flag.BoolVar(&test, testArg, true, "Testing")
	flagified := Flagify(testArg)
	os.Args = []string{flagified}
	flag.Parse()

	if !test {
		t.Errorf("Expected --test to be set, got these args: %v", os.Args)
	}

}
