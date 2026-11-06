//go:build linux

package tray

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// EnsureDesktopIntegrationWithBaseDir registers the icon and .desktop file under the specified base directory.
func EnsureDesktopIntegrationWithBaseDir(baseDir string, execPath string, iconBytes []byte) error {
	if len(iconBytes) == 0 {
		return nil
	}

	if execPath == "" {
		currentExec, err := os.Executable()
		if err == nil {
			execPath = currentExec
		} else {
			execPath = "applauncher3"
		}
	}

	// 1. Install hicolor 512x512 icon
	iconDir := filepath.Join(baseDir, "icons", "hicolor", "512x512", "apps")
	if err := os.MkdirAll(iconDir, 0755); err != nil {
		return fmt.Errorf("failed to create icon directory: %w", err)
	}
	iconPath := filepath.Join(iconDir, "applauncher3.png")

	writeIcon := true
	if existing, err := os.ReadFile(iconPath); err == nil && bytes.Equal(existing, iconBytes) {
		writeIcon = false
	}
	if writeIcon {
		if err := os.WriteFile(iconPath, iconBytes, 0644); err != nil {
			return fmt.Errorf("failed to write icon file: %w", err)
		}
	}

	// 2. Install desktop entry
	appsDir := filepath.Join(baseDir, "applications")
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		return fmt.Errorf("failed to create applications directory: %w", err)
	}
	desktopPath := filepath.Join(appsDir, "applauncher3.desktop")

	desktopContent := "[Desktop Entry]\n" +
		"Type=Application\n" +
		"Name=AppLauncher3\n" +
		"Comment=Sequential & Staggered Process Automator\n" +
		"Exec=" + execPath + "\n" +
		"Icon=applauncher3\n" +
		"Terminal=false\n" +
		"Categories=Utility;\n" +
		"StartupWMClass=applauncher3\n"

	writeDesktop := true
	if existing, err := os.ReadFile(desktopPath); err == nil && string(existing) == desktopContent {
		writeDesktop = false
	}
	if writeDesktop {
		if err := os.WriteFile(desktopPath, []byte(desktopContent), 0644); err != nil {
			return fmt.Errorf("failed to write desktop file: %w", err)
		}
		// Refresh system databases asynchronously
		go func() {
			_ = exec.Command("update-desktop-database", appsDir).Run()
			_ = exec.Command("kbuildsycoca6", "--noincremental").Run()
		}()
	}

	return nil
}

// EnsureDesktopIntegration guarantees that the application icon and .desktop file
// are registered in XDG user directories (~/.local/share/icons and ~/.local/share/applications)
// so that Wayland compositors (KWin, Mutter) display the window icon on the titlebar and dock.
func EnsureDesktopIntegration(iconBytes []byte) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	baseDir := filepath.Join(home, ".local", "share")
	return EnsureDesktopIntegrationWithBaseDir(baseDir, "", iconBytes)
}
