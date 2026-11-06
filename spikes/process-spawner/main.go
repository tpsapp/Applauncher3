package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: spawner <pidfile> <command> [args...]")
		os.Exit(1)
	}

	pidFile := os.Args[1]
	targetCmd := os.Args[2]
	var targetArgs []string
	if len(os.Args) > 3 {
		targetArgs = os.Args[3:]
	}

	cmd := exec.Command(targetCmd, targetArgs...)
	// Detach process: new process group, immune to parent terminal signals
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Detach standard streams so parent exit does not close broken pipes on child
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Spawn error: %v\n", err)
		os.Exit(1)
	}

	childPID := cmd.Process.Pid
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(childPID)), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write PID file: %v\n", err)
		os.Exit(1)
	}

	absPath, _ := filepath.Abs(pidFile)
	fmt.Printf("SPAWNED_CHILD_PID:%d PID_FILE:%s\n", childPID, absPath)
	// Parent exits immediately without waiting for child
	os.Exit(0)
}
