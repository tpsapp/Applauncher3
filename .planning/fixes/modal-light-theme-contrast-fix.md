# Bug Fix: Modal Card Contrast and Legibility in Light Theme

## Defect Summary
In light theme, portions of the "Add Application" and "Preferences & Launcher Settings" modals became unreadable. Grouped setting cards and the "Advanced Launch Conditions" panel rendered with dark charcoal backgrounds (`rgba(9, 9, 11, 0.6)`), causing dark foreground text to display on a dark background. In addition, disabled inputs became translucent muddy gray via `disabled:opacity-40`, and secondary action buttons lacked distinct borders.

## Root Cause Analysis
1. In `frontend/src/components/SettingsModal.svelte` (preference cards) and `frontend/src/components/ItemModal.svelte` (Advanced Launch Conditions container), elements use the Tailwind opacity-modifier class `bg-zinc-950/60`.
2. In CSS, `bg-zinc-950/60` translates to the literal class selector `.bg-zinc-950\/60`, which was not covered by the `.bg-zinc-950` override in `app.css`. As a result, Tailwind's default `rgba(9, 9, 11, 0.6)` applied on top of the modal's white surface, producing dark charcoal boxes (`#676767`).
3. Because typography rules properly transformed `.text-zinc-200` to `#0f172a` (near-black) and `.text-zinc-500` to `#64748b` (slate), the text rendered black-on-charcoal, making it partially or completely unreadable.
4. The `disabled:opacity-40` utility on the "Probe Timeout" input in `ItemModal.svelte` caused 40% opacity in light mode, making it look muddy rather than like a native disabled text input.
5. Secondary buttons (Cancel, Browse) were missing borders and contrast definition against light backgrounds.

## Target Files
- `frontend/src/app.css`: Added explicit `.light` overrides for `.bg-zinc-950\/60` (`#f8fafc`), borders (`#e2e8f0`), disabled inputs (`#f1f5f9` background, `#94a3b8` text, full opacity), placeholder colors, secondary buttons with borders, and high-contrast accent text colors (`text-indigo-400`, `text-amber-400`, `text-cyan-400`).
- `frontend/src/components/ItemModal.svelte`: Scoped `disabled:opacity-40` to `dark:disabled:opacity-40`.
- `frontend/tests/theme.test.mjs`: Added `TC_THEME_04_LightModeCSSRulesCoverage` asserting required `.light` CSS rules for modal cards and inputs.

## Regression Tests
`frontend/tests/theme.test.mjs`:
- `TC_THEME_04_LightModeCSSRulesCoverage`: Asserts coverage for `.bg-zinc-950\/60`, `.border-zinc-800\/80`, `input:disabled`, `button.bg-zinc-800`, and `select`.

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
▶ Modal Field Reset & Lifecycle Invariant Tests (3.10ms)
▶ Theme Resolution & DOM Synchronization Tests
  ✔ TC_THEME_01_ExplicitThemeResolution (0.58ms)
  ✔ TC_THEME_02_SystemThemeFollowsMatchMedia (0.19ms)
  ✔ TC_THEME_03_ApplyThemeTogglesClasses (0.19ms)
  ✔ TC_THEME_04_LightModeCSSRulesCoverage (2.77ms)
✔ Theme Resolution & DOM Synchronization Tests (4.49ms)
```

```bash
make all
```
Vite production build succeeded in 601ms; production binary compiled into `bin/applauncher3`.

## Side-Effect Audit
- [x] Public API contracts preserved.
- [x] No extraneous formatting or refactoring changes included.
- [x] Full existing test suite passes cleanly under race detection.
