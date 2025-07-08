package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	pf "github.com/grafana/nethax/pkg/probeflags"
)

const (
	exitCodeSuccess     = 0
	exitCodeFailure     = 1
	exitCodeConfigError = 2
)

var (
	url            string
	timeout        time.Duration
	expectedStatus int
	testType       string
	expectFail     bool
)

func main() {
	flag.StringVar(&url, pf.ArgURL, "", "URL or host:port to connect to")
	flag.DurationVar(&timeout, pf.ArgTimeout, 5*time.Second, "Timeout value (e.g. 5s, 1m)")
	flag.IntVar(&expectedStatus, pf.ArgExpectedStatus, 200, "Expected HTTP status code (0 for connection failure)")
	flag.StringVar(&testType, pf.ArgType, pf.TestTypeHTTP, "Type of test (http or tcp)")
	flag.BoolVar(&expectFail, pf.ArgExpectFail, false, "Whether the test is expected to fail (TCP and DNS tests only)")
	flag.Parse()

	if url == "" {
		fmt.Println("Error: URL must be specified")
		os.Exit(exitCodeFailure)
	}

	var probe Probe

	switch testType {
	case pf.TestTypeTCP:
		probe = NewTCPProbe(url, expectFail)
	case pf.TestTypeHTTP:
		probe = NewHTTPProbe(url, expectedStatus)
	case pf.TestTypeDNS:
		probe = NewDNSProbe(url, expectFail)
	default:
		fmt.Println("Error: Invalid test type: ", testType)
		os.Exit(exitCodeConfigError)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := probe.Run(ctx); err != nil {
		fmt.Println("Probe failed unexpectedly:", err)
		os.Exit(exitCodeFailure)
	}

	fmt.Println("Probe succeeded")
	os.Exit(exitCodeSuccess)
}
