package runner

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestProbeTCPPort_Success(t *testing.T) {
	// Start a local test TCP listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start listener: %v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = ProbeTCPPort(ctx, port, 1*time.Second, 20*time.Millisecond)
	if err != nil {
		t.Fatalf("Expected probe to succeed on open port %d, got: %v", port, err)
	}
}

func TestProbeTCPPort_Timeout(t *testing.T) {
	// Find an unused port and immediately close it
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = ProbeTCPPort(ctx, port, 100*time.Millisecond, 20*time.Millisecond)
	if !errors.Is(err, ErrProbeTimeout) {
		t.Fatalf("Expected ErrProbeTimeout, got: %v", err)
	}
}
