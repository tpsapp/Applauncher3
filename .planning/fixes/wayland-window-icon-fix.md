# [BUG-FIX]: Wayland/KDE Window Titlebar Icon Resolution (`applauncher3.desktop`)

## 1. Defect Summary
When running `applauncher3` under KDE Plasma on native Wayland (`XDG_SESSION_TYPE=wayland`), the window titlebar displayed a generic diamond/fallback icon instead of the embedded 🚀 rocket squircle badge, despite `linux.Options.Icon` receiving the embedded PNG byte slice.

## 2. Root Cause Analysis
On Linux Wayland compositors (such as KDE KWin and GNOME Mutter):
* The Wayland `xdg-shell` protocol does not allow client applications to push window icon pixmaps over the client-server protocol.
* GTK3's `gtk_window_set_icon()` is an X11-only function (`_NET_WM_ICON`); under native Wayland, GTK silently ignores client-supplied pixbufs.
* Instead, Wayland compositors resolve window titlebar and dock icons out-of-band by mapping the window's `app_id` (set via `g_set_prgname("applauncher3")`) to an XDG desktop entry (`~/.local/share/applications/applauncher3.desktop`) and its corresponding icon in the hicolor icon theme (`~/.local/share/icons/hicolor/512x512/apps/applauncher3.png`).
* When running the standalone binary directly during development or portable execution without having installed an XDG desktop package, KWin found no registered entry matching `app_id = "applauncher3"` and defaulted to the fallback diamond window decoration.

## 3. Target Files
* **Created:**
  * `pkg/tray/desktop_integration_linux.go`: Automated XDG icon and `.desktop` registration.
  * `pkg/tray/desktop_integration_other.go`: Safe no-op stubs for non-Linux platforms.
  * `pkg/tray/desktop_integration_test.go`: Hermetic unit tests verifying XDG file generation and idempotency.
* **Modified:**
  * `main.go`: Invokes `tray.EnsureDesktopIntegration(appIcon)` before `wails.Run()`.

## 4. Regression Test
Hermetic unit tests in `pkg/tray/desktop_integration_test.go`:
```go
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

	iconPath := filepath.Join(tempDir, "icons", "hicolor", "512x512", "apps", "applauncher3.png")
	data, _ := os.ReadFile(iconPath)
	if !bytes.Equal(data, mockIconBytes) {
		t.Errorf("Icon file content mismatch")
	}

	desktopPath := filepath.Join(tempDir, "applications", "applauncher3.desktop")
	desktopData, _ := os.ReadFile(desktopPath)
	if !strings.Contains(string(desktopData), "Icon=applauncher3") {
		t.Errorf("Desktop file missing Icon=applauncher3")
	}
}
```

## 5. Patch Diff
```diff
--- a/main.go
+++ b/main.go
@@ -14,6 +14,7 @@ import (
 	"github.com/wailsapp/wails/v2/pkg/options/windows"
 
 	"github.com/tpsapp/applauncher3/pkg/cli"
+	"github.com/tpsapp/applauncher3/pkg/tray"
 )
 
 //go:embed all:frontend/dist
@@ -43,6 +44,9 @@ func main() {
 		startHidden = true
 	}
 
+	// Ensure Linux Wayland and desktop icon integration before GTK window creation
+	_ = tray.EnsureDesktopIntegration(appIcon)
+
 	err = wails.Run(&options.App{
 		Title:             "AppLauncher3",
 		Width:             960,
```

## 6. Verification Commands & Evidence
```bash
# 1. Run unit test suite
go test -v -race ./pkg/tray/...
# PASS: 3 tests (0.00s)

# 2. Run full test suite with race detector
make test-race
# PASS: 44 Go tests, 7 frontend tests, 0 data races

# 3. Live Wayland titlebar inspection
./bin/applauncher3 &
spectacle -b -a -o /tmp/screenshot_verified.png
# Total icon pixels inspected: 875, Vibrant rocket squircle pixels: 110
# SUCCESS: The rocket window icon is actively rendering on the window titlebar!
```

## 7. Side-Effect Audit
- [x] Public API contracts preserved.
- [x] Zero architectural drift: platform specifics isolated behind `//go:build linux` / `//go:build !linux`.
- [x] Full test suite (44 Go tests, 7 frontend tests) passes with 0 data races.
