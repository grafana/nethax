package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	//RootCmd is the root level command that all other commands attach to
	RootCmd = &NethaxCommand{ // base command
		Command: &cobra.Command{
			Use:   "runner --help",
			Short: "nethax test runner",
		},
	}
)

func init() {
	addCommands()
}

type NethaxCommand struct {
	*cobra.Command
}

// AddCommand adds child commands and adds child commands for cobra as well.
func (c *NethaxCommand) AddCommand(commands ...*NethaxCommand) {
	for _, cmd := range commands {
		c.Command.AddCommand(cmd.Command)
	}
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
