//go:build !linux

package tray

// EnsureDesktopIntegrationWithBaseDir is a no-op on non-Linux platforms.
func EnsureDesktopIntegrationWithBaseDir(baseDir string, execPath string, iconBytes []byte) error {
	return nil
}

// EnsureDesktopIntegration is a no-op on non-Linux platforms.
func EnsureDesktopIntegration(iconBytes []byte) error {
	return nil
}
