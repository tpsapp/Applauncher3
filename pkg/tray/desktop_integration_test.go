package tray

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureDesktopIntegration_CreatesIconAndDesktopFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "desktop-integration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mockIconBytes := []byte("fake-png-icon-bytes-1234567890")
	execPath := "/custom/path/bin/applauncher3"

	err = EnsureDesktopIntegrationWithBaseDir(tempDir, execPath, mockIconBytes)
	if err != nil {
		t.Fatalf("EnsureDesktopIntegrationWithBaseDir returned error: %v", err)
	}

	// 1. Verify icon file
	iconPath := filepath.Join(tempDir, "icons", "hicolor", "512x512", "apps", "applauncher3.png")
	data, err := os.ReadFile(iconPath)
	if err != nil {
		t.Fatalf("Failed to read created icon file: %v", err)
	}
	if !bytes.Equal(data, mockIconBytes) {
		t.Errorf("Icon file content mismatch: got %s, want %s", string(data), string(mockIconBytes))
	}

	// 2. Verify desktop file
	desktopPath := filepath.Join(tempDir, "applications", "applauncher3.desktop")
	desktopData, err := os.ReadFile(desktopPath)
	if err != nil {
		t.Fatalf("Failed to read created desktop file: %v", err)
	}
	content := string(desktopData)
	if !strings.Contains(content, "Icon="+iconPath) {
		t.Errorf("Desktop file missing Icon=%s: %s", iconPath, content)
	}
	if !strings.Contains(content, "Exec="+execPath) {
		t.Errorf("Desktop file missing Exec=%s: %s", execPath, content)
	}
	if !strings.Contains(content, "StartupNotify=true") {
		t.Errorf("Desktop file missing StartupNotify=true: %s", content)
	}
	if !strings.Contains(content, "StartupWMClass=applauncher3") {
		t.Errorf("Desktop file missing StartupWMClass=applauncher3: %s", content)
	}

	// 3. Verify pixmaps fallback file
	pixmapPath := filepath.Join(tempDir, "pixmaps", "applauncher3.png")
	pixmapData, err := os.ReadFile(pixmapPath)
	if err != nil {
		t.Fatalf("Failed to read created pixmap file: %v", err)
	}
	if !bytes.Equal(pixmapData, mockIconBytes) {
		t.Errorf("Pixmap file content mismatch")
	}
}

func TestEnsureDesktopIntegration_Idempotent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "desktop-integration-idem-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mockIconBytes := []byte("fake-png-icon-bytes")
	execPath := "/bin/applauncher3"

	// Run first time
	if err := EnsureDesktopIntegrationWithBaseDir(tempDir, execPath, mockIconBytes); err != nil {
		t.Fatalf("First run failed: %v", err)
	}

	// Run second time (should be no-op without error)
	if err := EnsureDesktopIntegrationWithBaseDir(tempDir, execPath, mockIconBytes); err != nil {
		t.Fatalf("Second run failed: %v", err)
	}
}

func TestEnsureDesktopIntegration_EmptyBytesNoop(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "desktop-integration-empty-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := EnsureDesktopIntegrationWithBaseDir(tempDir, "/bin/applauncher3", nil); err != nil {
		t.Fatalf("Expected nil error for empty bytes, got: %v", err)
	}

	// Directory should be untouched
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries in tempDir, found %d", len(entries))
	}
}
