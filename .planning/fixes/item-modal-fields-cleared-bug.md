# Bug Fix: Item Modal Field Clearance on Add Item

## Defect Summary
When opening the "Add Application" modal after having previously entered or added an item, the modal input fields retained the data from the previous item rather than resetting to clean, blank defaults.

## Root Cause Analysis
In Svelte 5, `ItemModal.svelte` maintained reactive form fields using `$state` primitives and synchronized them via `$effect(() => { if (item) { ... } else { ... } })`.

However:
1. `ItemModal.svelte` is permanently mounted in `App.svelte` and conditionally displayed using `{#if isOpen}` inside its template.
2. The `$effect` in `ItemModal.svelte` only read `item`, establishing dependency tracking strictly on the `item` prop. It did not track `isOpen`.
3. When the user clicked "Add Application", `item` is `null`. The user entered information and saved or dismissed the modal (`isOpen` became `false`).
4. When the user subsequently clicked "Add Application" again, `item` was still `null`. Because `item` never changed (remaining `null`), Svelte's reactive engine did not re-execute `$effect`.
5. Consequently, the internal `$state` variables retained the strings and flags from the previous session.

In addition, `SettingsModal.svelte` had an analogous flaw where unsaved modifications were retained if the user closed and reopened settings, because its `$effect` only tracked `settings` without tracking `isOpen`.

## Target Files
- `frontend/src/components/ItemModal.svelte` (Added `if (isOpen)` guard to effect so reopening resets form state)
- `frontend/src/components/SettingsModal.svelte` (Added `isOpen && settings` condition to effect)
- `frontend/src/App.svelte` (Explicitly reset `editingItem = null` upon saving or closing modal)
- `frontend/package.json` (Added `test` script running `node --test tests/*.test.mjs`)
- `frontend/tests/modal_reset.test.mjs` (Regression test suite asserting field clearance and lifecycle invariants)
- `Makefile` (Linked frontend unit tests to `make test` and `make test-race`)

## Regression Tests
In `frontend/tests/modal_reset.test.mjs`:
- `TC_MODAL_01_AddItem_FieldsResetOnReopen_Patched`: Verifies that adding an item, typing field data, closing the modal, and reopening "Add Application" clears all fields (`name`, `path`, `args`, `workingDirectory`, `waitForExit`, `tcpPortProbe`) back to their clean defaults.
- `TC_MODAL_02_AddItem_AfterEditItemPopulatesAndClears`: Verifies that switching between Edit Mode (existing item) and Add Mode resets the form state accurately.
- `TC_MODAL_03_SettingsModal_ResetOnCancel`: Verifies that closing Settings without saving properly discards dirty changes on reopen.

## Verification Evidence
```bash
make test-race
```
```
go test -v -race -timeout 30s ./...
=== RUN   TestInferItemMetadata
--- PASS: TestInferItemMetadata (0.00s)
... (All 36 Go unit & integration tests pass with 0 race warnings) ...
go test -v -race tests/e2e/e2e_test.go
=== RUN   TestE2E_FullLifecycleExecution
--- PASS: TestE2E_FullLifecycleExecution (0.21s)
PASS
cd frontend && npm test
▶ Modal Field Reset & Lifecycle Invariant Tests
  ✔ TC_MODAL_01_AddItem_FieldsResetOnReopen_Patched (1.68ms)
  ✔ TC_MODAL_02_AddItem_AfterEditItemPopulatesAndClears (0.34ms)
  ✔ TC_MODAL_03_SettingsModal_ResetOnCancel (0.22ms)
✔ Modal Field Reset & Lifecycle Invariant Tests (2.86ms)
ℹ tests 3
ℹ suites 1
ℹ pass 3
ℹ fail 0
```

```bash
make all
```
Vite production frontend built in 573ms; production binary compiled cleanly into `bin/applauncher3`.

## Side-Effect Audit
- [x] Public API contracts preserved.
- [x] No extraneous formatting or refactoring changes included.
- [x] Full existing test suite passes cleanly under race detection.
