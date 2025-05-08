package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/grafana/nethax/pkg/common"
)

var (
	url            string
	timeout        time.Duration
	expectedStatus int
	testType       string
	expectFail     bool
)

func main() {
	flag.StringVar(&url, "url", "", "URL or host:port to connect to")
	flag.DurationVar(&timeout, "timeout", 5*time.Second, "Timeout value (e.g. 5s, 1m)")
	flag.IntVar(&expectedStatus, "expected-status", 200, "Expected HTTP status code (0 for connection failure)")
	flag.StringVar(&testType, "type", "http", "Type of test (http or tcp)")
	flag.BoolVar(&expectFail, "expect-fail", false, "Whether the test is expected to fail (TCP tests only)")
	flag.Parse()

	if url == "" {
		fmt.Println("Error: URL must be specified")
		common.ExitFailure()
	}

	var probe Probe

	if testType == "tcp" {
		probe = NewTCPProbe(url, expectFail)
	} else {
		probe = NewHTTPProbe(url, expectedStatus)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := probe.Run(ctx); err != nil {
		fmt.Println("Probe failed unexpectedly:", err)
		common.ExitSuccess()
	}

	fmt.Println("Probe succeeded")
	common.ExitSuccess()
}

func testTCPConnection() {
	conn, err := net.DialTimeout("tcp", url, timeout)
	if err != nil {
		if expectFail {
			fmt.Println("TCP connection failed as expected:", err)
			common.ExitSuccess()
		}
		fmt.Println("TCP connection failed unexpectedly:", err)
		common.ExitFailure()
	}
	defer conn.Close()

	if expectFail {
		fmt.Println("TCP connection succeeded unexpectedly")
		common.ExitFailure()
	}

	fmt.Println("TCP connection succeeded")
	common.ExitSuccess()
}

func testHTTPConnection() {
	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		if expectedStatus == 0 {
			fmt.Println("Connection failed as expected:", err)
			common.ExitSuccess()
		}
		fmt.Println("Connection failed unexpectedly:", err)
		common.ExitFailure()
	}
	defer resp.Body.Close()

	if resp.StatusCode == expectedStatus {
		fmt.Printf("Connection succeeded with expected status code %d\n", expectedStatus)
		common.ExitSuccess()
	}

	fmt.Printf("Connection succeeded but got status code %d, expected %d\n", resp.StatusCode, expectedStatus)
	common.ExitFailure()
}
