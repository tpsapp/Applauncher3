# Bug Fix: Theme Selection and Close-to-Tray Window Termination

## Defect Summary
1. Changes to the color theme setting (`system`, `dark`, `light`) in Preferences were persisted to `config.json` but never applied to the DOM or CSS; the frontend lacked theme handling and remained locked into dark mode.
2. Turning off "Close to System Tray" in Preferences was ignored upon closing the application. AppLauncher3 continued to hide into the background rather than terminating immediately, leaving orphaned background processes without any system tray icon on Linux KDE.

## Root Cause Analysis
1. **Theme Inaction**:
   - `frontend/src/App.svelte` and `index.html` hardcoded `.dark` and dark-mode Tailwind classes on `<html lang="en" class="dark">` and `<main>`.
   - Neither `onMount` nor `handleSaveSettings` invoked any theme-resolution logic or updated the HTML class list (`light` vs `dark`).
   - `frontend/src/app.css` lacked CSS rules for `.light` mode.

2. **Close-to-Tray / Window Termination Defect**:
   - In `main.go`, `HideWindowOnClose: true` was hardcoded inside `wails.Run(&options.App{ ... })`.
   - On Linux/WebKitGTK, `HideWindowOnClose: true` instructed GTK to directly bind `delete-event` to `gtk_widget_hide_on_delete`. This completely bypassed Wails' `Frontend.Quit()` and `OnBeforeClose` lifecycle callbacks.
   - Consequently, clicking the window close button (`X`) unconditionally hid the window on GTK regardless of the `CloseToTray: false` setting in `config.json`.
   - Furthermore, Wails v2 does not provide a native Linux system tray runtime, making a hidden window inaccessible to users on Linux desktop environments such as KDE Plasma.

## Target Files
- `main.go`: Set `HideWindowOnClose: false`, hooked `OnBeforeClose: app.beforeClose`, and loaded `cfg.Settings.StartMinimized`.
- `app.go`: Implemented `beforeClose(ctx)` lifecycle handler to inspect `CloseToTray`; permits immediate window termination when `false` and calls `WindowHide` when `true`. Cleanly cancels active queue engine on `shutdown(ctx)`.
- `app_test.go`: Added `TestBeforeClose_RespectsCloseToTray` verifying that `CloseToTray=false` returns `false` (immediate exit) and `CloseToTray=true` returns `true`.
- `pkg/models/config.go`: Default `CloseToTray` to `false` on Linux in `NewDefaultConfig()` to prevent trapping new Linux users.
- `frontend/src/theme.ts`: Created theme controller resolving `system`, `dark`, and `light` themes with `matchMedia` listeners and DOM class synchronization.
- `frontend/src/app.css`: Added complete `.light` theme styling across backgrounds, cards, typography, form inputs, buttons, and scrollbars.
- `frontend/src/App.svelte`: Integrated `theme.ts` reactivity on startup and during preferences save.
- `frontend/src/components/SettingsModal.svelte`: Clarified close-to-tray setting label.
- `frontend/tests/theme.test.mjs`: Unit tests for theme resolution, `matchMedia` tracking, and DOM class manipulation.

## Regression Tests
### Go Backend (`app_test.go`)
```go
func TestBeforeClose_RespectsCloseToTray(t *testing.T) {
	// With CloseToTray = false, beforeClose must return false (exit immediately)
	cfg.Settings.CloseToTray = false
	prevent := app.beforeClose(ctx)
	// assert prevent == false

	// With CloseToTray = true, beforeClose must return true (prevent termination)
	cfg.Settings.CloseToTray = true
	prevent = app.beforeClose(ctx)
	// assert prevent == true
}
```

### Frontend (`frontend/tests/theme.test.mjs`)
- `TC_THEME_01_ExplicitThemeResolution`: Asserts `dark` and `light` resolve correctly.
- `TC_THEME_02_SystemThemeFollowsMatchMedia`: Asserts `system` switches according to OS preference.
- `TC_THEME_03_ApplyThemeTogglesClasses`: Asserts `.light` and `.dark` classes are synchronized on `document.documentElement`.

## Verification Commands & Evidence
```bash
make test-race
```
```
=== RUN   TestInferItemMetadata
--- PASS: TestInferItemMetadata (0.00s)
=== RUN   TestBeforeClose_RespectsCloseToTray
--- PASS: TestBeforeClose_RespectsCloseToTray (0.00s)
... (All 37 Go tests passing with 0 race warnings) ...
▶ Modal Field Reset & Lifecycle Invariant Tests (3.03ms)
▶ Theme Resolution & DOM Synchronization Tests
  ✔ TC_THEME_01_ExplicitThemeResolution (0.58ms)
  ✔ TC_THEME_02_SystemThemeFollowsMatchMedia (0.18ms)
  ✔ TC_THEME_03_ApplyThemeTogglesClasses (0.20ms)
✔ Theme Resolution & DOM Synchronization Tests (1.62ms)
```

```bash
make all
```
Vite production build succeeded in 580ms; production binary compiled into `bin/applauncher3`.

## Side-Effect Audit
- [x] Public API contracts preserved.
- [x] No extraneous formatting or refactoring changes included.
- [x] Full existing test suite passes cleanly under race detection.
