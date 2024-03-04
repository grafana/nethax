package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Pod2Pod() *Command {
	return &Command{
		Command: &cobra.Command{
			Use: "pod2pod",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println("pod2pod not implemented yet")
			},
		},
	}
}
