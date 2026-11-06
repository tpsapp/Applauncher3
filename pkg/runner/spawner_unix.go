//go:build !windows

package runner

import (
	"os/exec"
	"syscall"

	"github.com/tpsapp/applauncher3/pkg/models"
)

type UnixSpawner struct{}

func NewOSSpawner() ProcessSpawner {
	return &UnixSpawner{}
}

func (s *UnixSpawner) Spawn(item *models.LaunchItem) (*SpawnResult, error) {
	if err := checkWorkingDirectory(item.WorkingDirectory); err != nil {
		return nil, err
	}

	cmd := exec.Command(item.Path, item.Arguments...)
	if item.WorkingDirectory != "" {
		cmd.Dir = item.WorkingDirectory
	}

	// Detached process group: immune to parent termination signals and SIGHUP
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Disconnect standard streams to prevent SIGPIPE when launcher closes
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &SpawnResult{
		PID:  cmd.Process.Pid,
		Wait: cmd.Wait,
	}, nil
}
