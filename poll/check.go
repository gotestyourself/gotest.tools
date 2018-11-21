package poll

import (
	"net"
	"os"
)

// Check is a function which will be used as check for the WaitOn method.
type Check func(t LogT) Result

// FileCheck looks on filesystem and check that the path exists.
func FileCheck(path string) Check {
	return func(t LogT) Result {
		_, err := os.Stat(path)
		if err != nil && os.IsNotExist(err) {
			t.Logf("waiting on file %s to be available", path)
			return Continue("file %s not available", path)
		}
		if err != nil {
			return Error(err)
		}

		return Success()
	}
}

// SocketCheck try to open a connection to the address on the
// named network.
func SocketCheck(network, address string) Check {
	return func(t LogT) Result {
		_, err := net.Dial(network, address)
		if err != nil {
			t.Logf("waiting on socket %s:%s to be available...", network, address)
			return Continue("socket %s:%s not available", network, address)
		}
		return Success()
	}
}
