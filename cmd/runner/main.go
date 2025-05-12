package main

import (
	"fmt"
	"strings"

	"github.com/grafana/nethax/pkg/common"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "runner --help",
		Short: "nethax test runner",
	}

	root.AddCommand(ExecuteTest())

	if err := root.Execute(); err != nil {
		if !strings.Contains(err.Error(), "unknown command") {
			fmt.Println(err)
		}
		common.ExitConfigError()
	}
}
