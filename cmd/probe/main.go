package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	port       int
	address    string
	timeout    int
	expectFail bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "Nethax Prober",
		Short: "Nethax Prober check ability to do connectivity",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Port: %s\n", port)
			fmt.Printf("Address: %s\n", address)
			fmt.Printf("timeout: %s\n", timeout)
			fmt.Printf("expectFail: %s\n", expectFail)
		},
	}

	// Printed value, shorthand, default values, description
	rootCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port number to use")
	rootCmd.Flags().StringVarP(&address, "address", "a", "", "Address to bind to")
	rootCmd.Flags().IntVarP(&timeout, "timeout", "t", 5, "Timeout value")
	rootCmd.Flags().BoolVarP(&expectFail, "expect failure", "f", false, "What parameter")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
