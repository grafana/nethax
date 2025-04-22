package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

var (
	url            string
	timeout        int
	expectedStatus int
	testType       string
	expectFail     bool
)

func main() {
	flag.StringVar(&url, "url", "", "URL or host:port to connect to")
	flag.IntVar(&timeout, "timeout", 5, "Timeout value in seconds")
	flag.IntVar(&expectedStatus, "expected-status", 200, "Expected HTTP status code (0 for connection failure)")
	flag.StringVar(&testType, "type", "http", "Type of test (http or tcp)")
	flag.BoolVar(&expectFail, "expect-fail", false, "Whether the test is expected to fail (TCP tests only)")
	flag.Parse()

	if url == "" {
		fmt.Println("Error: URL must be specified")
		os.Exit(1)
	}

	if testType == "tcp" {
		testTCPConnection()
	} else {
		testHTTPConnection()
	}
}

func testTCPConnection() {
	timeoutDuration := time.Duration(timeout) * time.Second
	conn, err := net.DialTimeout("tcp", url, timeoutDuration)
	if err != nil {
		if expectFail {
			fmt.Println("TCP connection failed as expected:", err)
			os.Exit(0)
		}
		fmt.Println("TCP connection failed unexpectedly:", err)
		os.Exit(1)
	}
	defer conn.Close()

	if expectFail {
		fmt.Println("TCP connection succeeded unexpectedly")
		os.Exit(1)
	}

	fmt.Println("TCP connection succeeded")
	os.Exit(0)
}

func testHTTPConnection() {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		if expectedStatus == 0 {
			fmt.Println("Connection failed as expected:", err)
			os.Exit(0)
		}
		fmt.Println("Connection failed unexpectedly:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == expectedStatus {
		fmt.Printf("Connection succeeded with expected status code %d\n", expectedStatus)
		os.Exit(0)
	}

	fmt.Printf("Connection succeeded but got status code %d, expected %d\n", resp.StatusCode, expectedStatus)
	os.Exit(1)
}
