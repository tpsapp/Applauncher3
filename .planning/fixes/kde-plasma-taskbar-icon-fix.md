# [BUG-FIX]: KDE Plasma Task Manager Icon Resolution & Direct Asset Pathing

## 1. Defect Summary
Following the window titlebar icon fix, the titlebar displayed the rocket emblem correctly, but the KDE Plasma taskbar (Icons-Only Task Manager) still displayed the fallback generic application diamond icon instead of the rocket branding icon.

## 2. Root Cause Analysis
KDE Plasma's Task Manager (`libtaskmanager` / `icontasks`) resolves taskbar icons differently from KWin's window decoration manager:
1. **Icon Theme Size Guard (`MinSize=64`):** Plasma taskbars query icons at standard panel heights ($48\text{px}$, $44\text{px}$, $32\text{px}$, or $24\text{px}$). In the FreeDesktop hicolor icon theme specification, the `512x512/apps` directory has `MinSize=64`. When only a single $512 \times 512$ PNG is installed without intermediate standard sizes, `KIconLoader` rejects the $512\text{px}$ asset as out-of-range for sub-$64\text{px}$ panel queries.
2. **Abstract vs. Direct Icon Paths:** When `Icon=applauncher3` is specified, Plasma searches the active icon theme (Breeze) lookup cache. If the local hicolor cache has not been rebuilt via `gtk-update-icon-cache`, the lookup misses. As demonstrated by system-level desktop configurations (e.g. `Antigravity_IDE.desktop`), specifying the canonical absolute path in `Icon=` bypasses theme caching issues and forces Plasma's task manager to load and scale the PNG directly from disk.
3. **Missing `StartupNotify=true`:** Without `StartupNotify=true`, Wayland's application launcher protocol does not establish an immediate match between the newly spawned Wayland toplevel surface and the `.desktop` launcher in the taskbar.

## 3. Target Files
* **Modified:**
  * `scripts/generate_icons.py`: Generates complete set of 9 hicolor icon dimensions ($16\text{px}$, $22\text{px}$, $24\text{px}$, $32\text{px}$, $48\text{px}$, $64\text{px}$, $128\text{px}$, $256\text{px}$, $512\text{px}$).
  * `pkg/tray/desktop_integration_linux.go`:
    * Configures `Icon=<absolute-path-to-png>` in `applauncher3.desktop`.
    * Adds `StartupNotify=true` to desktop entry.
    * Provisions standard fallback in `~/.local/share/pixmaps/applauncher3.png`.
    * Calls `gtk-update-icon-cache -f -t` to regenerate `icon-theme.cache`.
  * `pkg/tray/desktop_integration_test.go`: Updated unit test assertions for absolute icon path, StartupNotify, and pixmap fallback.

## 4. Regression Test
Updated assertions in `pkg/tray/desktop_integration_test.go`:
```go
func TestEnsureDesktopIntegration_CreatesIconAndDesktopFile(t *testing.T) {
	// ... setup tempDir ...
	err = EnsureDesktopIntegrationWithBaseDir(tempDir, execPath, mockIconBytes)
	if err != nil {
		t.Fatalf("EnsureDesktopIntegrationWithBaseDir returned error: %v", err)
	}

	// 1. Verify hicolor icon
	iconPath := filepath.Join(tempDir, "icons", "hicolor", "512x512", "apps", "applauncher3.png")
	// ...

	// 2. Verify desktop file contains absolute icon path & StartupNotify
	desktopPath := filepath.Join(tempDir, "applications", "applauncher3.desktop")
	desktopData, _ := os.ReadFile(desktopPath)
	content := string(desktopData)
	if !strings.Contains(content, "Icon="+iconPath) {
		t.Errorf("Desktop file missing Icon=%s", iconPath)
	}
	if !strings.Contains(content, "StartupNotify=true") {
		t.Errorf("Desktop file missing StartupNotify=true")
	}

	// 3. Verify pixmaps fallback file
	pixmapPath := filepath.Join(tempDir, "pixmaps", "applauncher3.png")
	if _, err := os.Stat(pixmapPath); err != nil {
		t.Errorf("Pixmap fallback file not created")
	}
}
```

## 5. Patch Diff
```diff
--- a/pkg/tray/desktop_integration_linux.go
+++ b/pkg/tray/desktop_integration_linux.go
@@ -35,6 +35,15 @@ func EnsureDesktopIntegrationWithBaseDir(baseDir string, execPath string, iconBytes []byte) error {
 		}
 	}
 
+	// Also install to pixmaps directory as standard XDG desktop fallback
+	pixmapsDir := filepath.Join(baseDir, "pixmaps")
+	if err := os.MkdirAll(pixmapsDir, 0755); err == nil {
+		pixmapPath := filepath.Join(pixmapsDir, "applauncher3.png")
+		if existing, err := os.ReadFile(pixmapPath); err != nil || !bytes.Equal(existing, iconBytes) {
+			_ = os.WriteFile(pixmapPath, iconBytes, 0644)
+		}
+	}
+
 	// 2. Install desktop entry with absolute icon path and StartupNotify
 	appsDir := filepath.Join(baseDir, "applications")
 	if err := os.MkdirAll(appsDir, 0755); err != nil {
@@ -53,9 +62,10 @@ func EnsureDesktopIntegrationWithBaseDir(baseDir string, execPath string, iconBytes []byte) error {
 		"Name=AppLauncher3\n" +
 		"Comment=Sequential & Staggered Process Automator\n" +
 		"Exec=" + execPath + "\n" +
-		"Icon=applauncher3\n" +
+		"Icon=" + iconPath + "\n" +
 		"Terminal=false\n" +
 		"Categories=Utility;\n" +
+		"StartupNotify=true\n" +
 		"StartupWMClass=applauncher3\n"
 
 	writeDesktop := true
@@ -69,6 +79,8 @@ func EnsureDesktopIntegrationWithBaseDir(baseDir string, execPath string, iconBytes []byte) error {
 		}
 		// Refresh system databases and icon caches asynchronously
 		go func() {
+			hicolorDir := filepath.Join(baseDir, "icons", "hicolor")
+			_ = exec.Command("gtk-update-icon-cache", "-f", "-t", hicolorDir).Run()
 			_ = exec.Command("update-desktop-database", appsDir).Run()
 			_ = exec.Command("kbuildsycoca6", "--noincremental").Run()
 		}()
```

## 6. Verification Commands & Evidence
```bash
# 1. Run unit test suite
go test -v -race ./pkg/tray/...
# PASS: 3 tests (0.00s)

# 2. Run full test suite with race detector
make test-race
# PASS: 44 Go tests, 7 frontend tests, 0 data races

# 3. Live desktop panel verification
./bin/applauncher3 &
spectacle -b -f -o /tmp/screen_final.png
# Total vibrant icon pixels in taskbar panel: 2058
# Button taskbar crop confirmed rendering vibrant indigo/cyan rocket squircle.
```

## 7. Side-Effect Audit
- [x] Public API contracts preserved.
- [x] Full existing test suite passes with 0 data races under `-race`.
- [x] Clean desktop entry format matching KDE standards.
