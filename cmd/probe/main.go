package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	port       int
	endpoint   string
	timeout    int
	expectFail bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "Nethax Prober",
		Short: "Nethax Prober check ability to do connectivity",
		Run:   testConnectivity,
	}

	// Printed value, shorthand, default values, description
	rootCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port number to connect to")
	rootCmd.Flags().StringVarP(&endpoint, "endpoint", "e", "localhost", "Endpoint to connect to")
	rootCmd.Flags().IntVarP(&timeout, "timeout", "t", 5, "Timeout value")
	rootCmd.Flags().BoolVarP(&expectFail, "expect-failure", "f", false, "Whether we expect the connection to fail or succeed")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func testConnectivity(cmd *cobra.Command, args []string) {
	address := fmt.Sprintf("%s:%d", endpoint, port)
	conn, err := net.DialTimeout("tcp", address, time.Duration(timeout)*time.Second)
	if err != nil {
		fmt.Println("Encountered an error trying to create network connection", err)
		if expectFail {
			fmt.Println("Failure expected, exiting cleanly.")
			os.Exit(0)
		} else {
			fmt.Println("Failure not expected, exiting with error.")
			os.Exit(1)
		}
	}

	fmt.Println("Connection succeeded to:", conn.RemoteAddr().String(), fmt.Sprintf("(%s)", address))
	err = conn.Close()
	if err != nil {
		fmt.Println("Encountered error trying to close TCP connection:", err)
		os.Exit(1)
	}
	os.Exit(0)
}
