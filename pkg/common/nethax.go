package common

func ExitNethax(exitStatus int, expectFail bool) int {
	if expectFail { // we expect connections to fail
		if exitStatus == 0 {
			exitStatus = 1 // connection succeeded when expected to fail
		} else {
			exitStatus = 0 // connection failed when expected to fail
		}
	}
	if exitStatus > 1 {
		exitStatus = 1 // normalize other failed exit codes to 1
	}

	return exitStatus
}
