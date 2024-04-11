package cmd

import (
	"github.com/grafana/nethax/interactive"
	"github.com/spf13/cobra"
)

func Interactive() *Command {
	p := interactive.Start()
	cmd := &Command{
		Command: &cobra.Command{
			Use:   "interactive",
			Short: "Run nethax in interactive mode",
			Run: func(cmd *cobra.Command, args []string) {
				interactive.Run(p)
			},
		},
		program: p,
	}
	return cmd
}
