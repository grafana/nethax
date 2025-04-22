package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	//RootCmd is the root level command that all other commands attach to
	RootCmd = &cobra.Command{
		Use:   "runner --help",
		Short: "nethax test runner",
	}
)

func init() {
	addCommands()
}

func Execute(args []string) error {
	if err := RootCmd.Execute(); err != nil {
		if !strings.Contains(err.Error(), "unknown command") {
			fmt.Println(err)
		}
		os.Exit(2)
	}

	return nil
}

func addCommands() {
	RootCmd.AddCommand(ExecuteTest())
}
