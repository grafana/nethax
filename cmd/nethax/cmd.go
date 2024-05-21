package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	//RootCmd is the root level command that all other commands attach to
	RootCmd = &Command{ // base command
		Command: &cobra.Command{
			Use:   "nethax",
			Short: "nethax!!! TODO",
		},
	}
)

func init() {
	addCommands()
}

type Command struct {
	*cobra.Command
}

func addSharedFlags(cmd *Command) {
	cmd.Flags().Int("timeout", 5, "Timeout for connections. Socket must connect successfully before this deadline elapses.")
	cmd.Flags().Bool("expect-fail", false, "Exit 0 on connection failure. Useful for tests where connections are expected to fail.")
}

// AddCommand adds child commands and adds child commands for cobra as well.
func (c *Command) AddCommand(commands ...*Command) {
	for _, cmd := range commands {
		addSharedFlags(cmd)
		c.Command.AddCommand(cmd.Command)
	}
}

func Execute(args []string) error {
	if err := RootCmd.Execute(); err != nil {
		if !strings.Contains(err.Error(), "unknown command") {
			fmt.Println(err)
		}
		os.Exit(-1)
	}

	return nil
}

func addCommands() {
	RootCmd.AddCommand(Pod2Pod())
	RootCmd.AddCommand(Pod2Remote())
}
