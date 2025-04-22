package common

import (
	"os"
)

const (
	ExitCodeSuccess     = 0
	ExitCodeFailure     = 1
	ExitCodeConfigError = 2
	ExitCodeNethaxError = 3
)

// ExitSuccess exits with code 0 (success)
func ExitSuccess() {
	os.Exit(ExitCodeSuccess)
}

// ExitFailure exits with code 1 (failure)
func ExitFailure() {
	os.Exit(ExitCodeFailure)
}

// ExitConfigError exits with code 2 (config error)
func ExitConfigError() {
	os.Exit(ExitCodeConfigError)
}

// ExitNethaxError exits with code 3 (nethax error)
func ExitNethaxError() {
	os.Exit(ExitCodeNethaxError)
}
