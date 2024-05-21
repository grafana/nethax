package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	// TODO: FLAGS
	timeout := 5 * time.Second
	host := "grafana.com"
	port := "443"
	expectFail := false

	// Make TCP connection
	address := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	// Handle connection failure
	if err != nil {
		fmt.Println("Encountered an error trying to create network connection", err)
		// Exit 0 if the connection had the expected result
		// note: sometimes, failure to connect is the expected result
		if expectFail {
			os.Exit(0)
		} else {
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
