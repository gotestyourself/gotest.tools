package poll

import (
	"context"
	"fmt"
	"net"
	"os"
)

// FileExists looks on filesystem and check that path exists.
func FileExists(path string) Check {
	return func(_ context.Context) error {
		_, err := os.Stat(path)
		switch {
		case os.IsNotExist(err):
			return Continue(fmt.Errorf("file %s does not exist", path))
		case err != nil:
			return err
		default:
			return nil
		}
	}
}

// Connection try to open a connection to the address on the
// named network. See [net.Dial] for a description of the network and
// address parameters.
func Connection(network, address string) Check {
	return func(ctx context.Context) error {
		var d net.Dialer
		_, err := d.DialContext(ctx, network, address)
		if err != nil {
			return Continue(fmt.Errorf("socket %s://%s not available", network, address))
		}
		return nil
	}
}
