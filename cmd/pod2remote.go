package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Pod2Remote() *Command {
	return &Command{
		Command: &cobra.Command{
			Use: "pod2remote",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println("pod2remote not implemented yet")
			},
		},
	}
}
