//go:build windows

package runner

import (
	"os/exec"
	"syscall"

	"github.com/tpsapp/applauncher3/pkg/models"
)

type WindowsSpawner struct{}

func NewOSSpawner() ProcessSpawner {
	return &WindowsSpawner{}
}

func (s *WindowsSpawner) Spawn(item *models.LaunchItem) (int, error) {
	if err := checkWorkingDirectory(item.WorkingDirectory); err != nil {
		return 0, err
	}

	cmd := exec.Command(item.Path, item.Arguments...)
	if item.WorkingDirectory != "" {
		cmd.Dir = item.WorkingDirectory
	}

	sysProcAttr := &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | 0x00000008, // DETACHED_PROCESS
	}

	switch item.LaunchState {
	case models.WindowHidden:
		sysProcAttr.HideWindow = true
		sysProcAttr.ShowWindow = syscall.SW_HIDE
	case models.WindowMinimized:
		sysProcAttr.ShowWindow = syscall.SW_SHOWMINIMIZED
	default:
		sysProcAttr.ShowWindow = syscall.SW_SHOWNORMAL
	}

	cmd.SysProcAttr = sysProcAttr
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return 0, err
	}

	return cmd.Process.Pid, nil
}
