package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/grafana/nethax/pkg/common"
	pf "github.com/grafana/nethax/pkg/probeflags"
)

var (
	url            string
	timeout        time.Duration
	expectedStatus int
	testType       string
	expectFail     bool
)

func main() {
	flag.StringVar(&url, pf.ArgUrl, "", "URL or host:port to connect to")
	flag.DurationVar(&timeout, pf.ArgTimeout, 5*time.Second, "Timeout value (e.g. 5s, 1m)")
	flag.IntVar(&expectedStatus, pf.ArgExpectedStatus, 200, "Expected HTTP status code (0 for connection failure)")
	flag.StringVar(&testType, pf.ArgType, pf.TestTypeHTTP, "Type of test (http or tcp)")
	flag.BoolVar(&expectFail, pf.ArgExpectFail, false, "Whether the test is expected to fail (TCP tests only)")
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
