package runner

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

var ErrProbeTimeout = errors.New("TCP port health check probe timed out")

// ProbeTCPPort continuously attempts to connect to 127.0.0.1:<port> until it accepts a TCP connection,
// or the specified timeout / context cancellation occurs.
func ProbeTCPPort(ctx context.Context, port int, timeout time.Duration, interval time.Duration) error {
	if port <= 0 || port > 65535 {
		return errors.New("invalid port number")
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	deadline := time.Now().Add(timeout)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%w: %s failed to respond within %v", ErrProbeTimeout, addr, timeout)
		}

		dialer := net.Dialer{Timeout: 200 * time.Millisecond}
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err == nil {
			_ = conn.Close()
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}
