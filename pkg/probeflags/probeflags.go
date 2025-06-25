package probeflags

import "fmt"

// Probe arguments -- can be Flagified
const (
	ArgURL            = "url"
	ArgTimeout        = "timeout"
	ArgExpectedStatus = "expected-status"
	ArgExpectFail     = "expect-fail"
	ArgType           = "type"
)

func Flagify(flag string) string {
	return fmt.Sprintf("--%s", flag)
}

const (
	TestTypeTCP  = "tcp"
	TestTypeHTTP = "http"
	TestTypeDNS  = "dns"
)
