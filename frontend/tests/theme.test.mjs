import { test, describe, beforeEach, afterEach } from 'node:test';
import assert from 'node:assert/strict';
import { getResolvedTheme, applyTheme, setupSystemThemeListener } from '../src/theme.ts';

describe('Theme Resolution & DOM Synchronization Tests', () => {
  let mockClassList;
  let originalDocument;
  let originalWindow;
  let systemPrefersDark = false;

  beforeEach(() => {
    mockClassList = new Set();

    originalDocument = globalThis.document;
    originalWindow = globalThis.window;

    globalThis.document = {
      documentElement: {
        classList: {
          add: (cls) => mockClassList.add(cls),
          remove: (cls) => mockClassList.delete(cls),
          contains: (cls) => mockClassList.has(cls),
        },
      },
    };

    globalThis.window = {
      matchMedia: (query) => ({
        matches: systemPrefersDark,
        addEventListener: () => {},
        removeEventListener: () => {},
      }),
    };
  });

  afterEach(() => {
    globalThis.document = originalDocument;
    globalThis.window = originalWindow;
  });

  test('TC_THEME_01_ExplicitThemeResolution', () => {
    assert.equal(getResolvedTheme('dark'), 'dark');
    assert.equal(getResolvedTheme('light'), 'light');
  });

  test('TC_THEME_02_SystemThemeFollowsMatchMedia', () => {
    systemPrefersDark = false;
    assert.equal(getResolvedTheme('system'), 'light');

    systemPrefersDark = true;
    assert.equal(getResolvedTheme('system'), 'dark');
  });

  test('TC_THEME_03_ApplyThemeTogglesClasses', () => {
    applyTheme('dark');
    assert.equal(mockClassList.has('dark'), true);
    assert.equal(mockClassList.has('light'), false);

    applyTheme('light');
    assert.equal(mockClassList.has('dark'), false);
    assert.equal(mockClassList.has('light'), true);

    systemPrefersDark = true;
    applyTheme('system');
    assert.equal(mockClassList.has('dark'), true);
    assert.equal(mockClassList.has('light'), false);
  });

  test('TC_THEME_04_LightModeCSSRulesCoverage', async () => {
    const fs = await import('node:fs/promises');
    const path = await import('node:path');
    const cssPath = path.resolve('src/app.css');
    const css = await fs.readFile(cssPath, 'utf-8');

    assert.ok(css.includes('html.light .bg-zinc-950\\/60'), 'CSS must style .bg-zinc-950/60 for option cards');
    assert.ok(css.includes('html.light .border-zinc-800\\/80'), 'CSS must style .border-zinc-800/80');
    assert.ok(css.includes('html.light input:disabled'), 'CSS must style disabled inputs cleanly');
    assert.ok(css.includes('html.light select'), 'CSS must style dropdown select for light mode');
    assert.ok(css.includes('html.light button.bg-zinc-800'), 'CSS must style secondary buttons with borders');
  });
});
