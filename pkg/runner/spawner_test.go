package runner

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/tpsapp/applauncher3/pkg/models"
)

func TestTC_SPW_01_HappyPath_Spawn(t *testing.T) {
	spawner := NewOSSpawner()
	item := &models.LaunchItem{
		Path:      "/bin/sh",
		Arguments: []string{"-c", "sleep 0.1"},
	}

	pid, err := spawner.Spawn(item)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}
	if pid <= 0 {
		t.Errorf("Expected positive PID, got %d", pid)
	}
}

func TestTC_SPW_02_DetachedOrphan_Survival(t *testing.T) {
	spawner := NewOSSpawner()
	item := &models.LaunchItem{
		Path:      "/bin/sh",
		Arguments: []string{"-c", "sleep 2"},
	}

	pid, err := spawner.Spawn(item)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Verify child process is active and running independently
	process, err := os.FindProcess(pid)
	if err != nil {
		t.Fatalf("Failed to find process: %v", err)
	}

	// Sending signal 0 checks process existence without killing it
	if err := process.Signal(syscall.Signal(0)); err != nil {
		t.Errorf("Process %d is not running: %v", pid, err)
	}

	// Cleanup
	_ = process.Kill()
}

func TestTC_SPW_03_MissingBinary(t *testing.T) {
	spawner := NewOSSpawner()
	item := &models.LaunchItem{
		Path: "/nonexistent/binary_path_xyz_12345",
	}

	_, err := spawner.Spawn(item)
	if err == nil {
		t.Fatalf("Expected error spawning nonexistent binary")
	}
}

func TestTC_SPW_04_PermissionDenied(t *testing.T) {
	tmpDir := t.TempDir()
	nonExecFile := filepath.Join(tmpDir, "not_executable.sh")
	if err := os.WriteFile(nonExecFile, []byte("#!/bin/sh\necho hi"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	spawner := NewOSSpawner()
	item := &models.LaunchItem{
		Path: nonExecFile,
	}

	_, err := spawner.Spawn(item)
	if err == nil {
		t.Fatalf("Expected permission error when running non-executable file")
	}
}

func TestTC_SPW_05_CustomWorkingDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	markerFile := filepath.Join(tmpDir, "marker.txt")

	spawner := NewOSSpawner()
	item := &models.LaunchItem{
		Path:             "/bin/sh",
		Arguments:        []string{"-c", "touch marker.txt"},
		WorkingDirectory: tmpDir,
	}

	pid, err := spawner.Spawn(item)
	if err != nil {
		t.Fatalf("Spawn failed: %v", err)
	}

	// Wait briefly for touch to execute
	time.Sleep(100 * time.Millisecond)

	if _, err := os.Stat(markerFile); os.IsNotExist(err) {
		t.Errorf("Expected marker.txt in working directory %s for PID %d", tmpDir, pid)
	}
}

func TestTC_SPW_06_InvalidWorkingDirectory(t *testing.T) {
	spawner := NewOSSpawner()
	item := &models.LaunchItem{
		Path:             "/bin/sh",
		Arguments:        []string{"-c", "echo hello"},
		WorkingDirectory: "/nonexistent/working_directory_xyz",
	}

	_, err := spawner.Spawn(item)
	if !errors.Is(err, ErrDirectoryNotFound) {
		t.Fatalf("Expected ErrDirectoryNotFound, got: %v", err)
	}
}
