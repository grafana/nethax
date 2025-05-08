package main

import (
	"context"
	"flag"
	"fmt"
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
		common.ExitFailure()
	}

	fmt.Println("Probe succeeded")
	common.ExitSuccess()
}
