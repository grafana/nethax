package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const (
	exitCodeFailure     = 1
	exitCodeConfigError = 2
)

func main() {
	root := &cobra.Command{
		Use:   "nethax-runner --help",
		Short: "nethax test runner",
	}

	root.AddCommand(ExecuteTest())

	if err := root.Execute(); err != nil {
		if !strings.Contains(err.Error(), "unknown command") {
			fmt.Println(err)
		}
		os.Exit(exitCodeConfigError)
	}
}
