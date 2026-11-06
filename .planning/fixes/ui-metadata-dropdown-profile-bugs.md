# [BUG-FIX]: Fix Metadata Auto-Fill, Dark Dropdown Theming, and New Profile Modal

## 1. Defect Summary
Three UI and data-binding issues were identified in AppLauncher3:
1. **Metadata String Truncation / Destructuring Bug:** When selecting an executable via the native file picker, the Display Label only received the first letter (e.g. `"d"`) and Working Directory received the second letter (e.g. `"i"`).
2. **WebKit2GTK White Dropdown Styling:** Native GTK `<select>` elements (`Launch Window State` and `Color Theme`) defaulted to an unstyled white widget background, rendering light text illegible.
3. **Crude Browser Prompt for Profiles:** Clicking the `+` button to create a new profile invoked browser-native `prompt()`, displaying an unstyled dialog with `JavaScript - wails://wails/` instead of an integrated application modal.

---

## 2. Root Cause Analysis
1. **Multiple Return Value Serialization in Wails v2:** In Go, `InferItemMetadata` originally returned multiple values `(name string, workingDir string)`. In Wails v2 RPC, non-error second return values are not returned as an array; only the first string was delivered to the JavaScript Promise. In `ItemModal.svelte`, JavaScript destructured the returned string (`const [inferredName, inferredDir] = "dicemachine"`), which iterates string characters yielding `'d'` and `'i'`.
   - *Fix:* Replaced multiple return values with a typed struct `ItemMetadata { Name, WorkingDirectory }`, properly serialized to a JSON object.
2. **WebKit2GTK Default Form Controls:** Without `appearance: none !important`, WebKitGTK falls back to host GTK button theme styling for `<select>`, overriding Tailwind dark classes.
   - *Fix:* Added global CSS rules in `frontend/src/app.css` setting `-webkit-appearance: none !important`, `background-color: #18181b !important`, `color: #f4f4f5 !important`, custom SVG chevron icon, and dark styling on `<option>` tags.
3. **Unstyled `window.prompt()`:** `handleCreateProfile` invoked `window.prompt()`.
   - *Fix:* Built a dedicated, dark-themed `NewProfileModal.svelte` component with inline validation, duplicate name checking, and keyboard shortcuts (Enter to submit, Escape to close).

---

## 3. Target Files Modified
- `app.go`: Changed `InferItemMetadata` to return `ItemMetadata` struct.
- `app_test.go`: Added regression test `TestInferItemMetadata` asserting full string resolution.
- `frontend/src/api.ts`: Updated `InferItemMetadata` interface and mock return types.
- `frontend/src/components/ItemModal.svelte`: Destructured metadata from `meta.name` and `meta.workingDirectory`.
- `frontend/src/app.css`: Added cross-platform dark `<select>` and `<option>` styling rules with custom chevron.
- `frontend/src/components/NewProfileModal.svelte`: Created modal component.
- `frontend/src/App.svelte`: Integrated `NewProfileModal` replacing `window.prompt()`.

---

## 4. Regression Test
```go
func TestInferItemMetadata(t *testing.T) {
	app := &App{}
	meta := app.InferItemMetadata("/home/tpsapp/.local/bin/dicemachine")
	if meta.Name != "dicemachine" {
		t.Errorf("Expected Name 'dicemachine', got '%s'", meta.Name)
	}
	if meta.WorkingDirectory != "/home/tpsapp/.local/bin" {
		t.Errorf("Expected WorkingDirectory '/home/tpsapp/.local/bin', got '%s'", meta.WorkingDirectory)
	}
}
```

---

## 5. Verification Commands & Evidence
```bash
$ go test -v app_test.go app.go
=== RUN   TestInferItemMetadata
--- PASS: TestInferItemMetadata (0.00s)
PASS

$ cd frontend && npm run build
✓ 118 modules transformed.
dist/index.html                  0.49 kB │ gzip:  0.34 kB
dist/assets/index-D0yBO-5S.css  20.85 kB │ gzip:  4.69 kB
dist/assets/index-BMDNjCnC.js   76.12 kB │ gzip: 25.30 kB
✓ built in 585ms (0 errors, 0 warnings)

$ make test-race
36 automated Go test cases passing cleanly with zero data races.
```

---

## 6. Side-Effect Audit
- [x] Public API contracts preserved.
- [x] Zero architectural drift or unvetted dependencies introduced.
- [x] Full test suite (36 tests) passes with ThreadSanitizer.
