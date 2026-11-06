import { test, describe } from 'node:test';
import assert from 'node:assert/strict';
import { tick } from 'svelte';
import { state, set, get, user_effect, effect_root } from 'svelte/internal/client';

describe('Modal Field Reset & Lifecycle Invariant Tests', () => {
  test('TC_MODAL_01_AddItem_FieldsResetOnReopen_Patched', async () => {
    let isOpen = state(false);
    let item = state(null);

    let name = state('');
    let path = state('');
    let args = state('');
    let workingDirectory = state('');
    let delaySeconds = state(2.0);
    let launchState = state('Normal');
    let enabled = state(true);
    let waitForExit = state(false);
    let tcpPortProbe = state(undefined);
    let probeTimeoutSeconds = state(10.0);

    // Patched reactive effect: tracks isOpen so reopening modal resets fields when item is null
    const cleanup = effect_root(() => {
      user_effect(() => {
        if (get(isOpen)) {
          const it = get(item);
          if (it) {
            set(name, it.name || '');
            set(path, it.path || '');
            set(args, (it.arguments || []).join(' '));
            set(workingDirectory, it.workingDirectory || '');
            set(delaySeconds, it.delaySeconds ?? 2.0);
            set(launchState, it.launchState || 'Normal');
            set(enabled, it.enabled ?? true);
            set(waitForExit, it.waitForExit ?? false);
            set(tcpPortProbe, it.tcpPortProbe || undefined);
            set(probeTimeoutSeconds, it.probeTimeoutSeconds ?? 10.0);
          } else {
            set(name, '');
            set(path, '');
            set(args, '');
            set(workingDirectory, '');
            set(delaySeconds, 2.0);
            set(launchState, 'Normal');
            set(enabled, true);
            set(waitForExit, false);
            set(tcpPortProbe, undefined);
            set(probeTimeoutSeconds, 10.0);
          }
        }
      });
    });

    await tick();

    // 1. First Add Item click
    set(isOpen, true);
    set(item, null);
    await tick();

    // 2. User inputs values into fields
    set(name, 'Firefox');
    set(path, '/usr/bin/firefox');
    set(args, '--private-window');
    set(workingDirectory, '/home/user');
    set(delaySeconds, 5.0);
    set(launchState, 'Maximized');
    set(enabled, true);
    set(waitForExit, true);
    set(tcpPortProbe, 8080);
    set(probeTimeoutSeconds, 15.0);
    await tick();

    assert.equal(get(name), 'Firefox');
    assert.equal(get(path), '/usr/bin/firefox');

    // 3. User saves item and modal closes
    set(isOpen, false);
    await tick();

    // 4. User clicks Add Item again (item is still null, isOpen becomes true)
    set(isOpen, true);
    await tick();

    // With the patch, fields MUST be cleared!
    assert.equal(get(name), '', 'name must be cleared on reopening Add Item');
    assert.equal(get(path), '', 'path must be cleared on reopening Add Item');
    assert.equal(get(args), '', 'args must be cleared on reopening Add Item');
    assert.equal(get(workingDirectory), '', 'workingDirectory must be cleared');
    assert.equal(get(delaySeconds), 2.0, 'delaySeconds must reset to 2.0');
    assert.equal(get(launchState), 'Normal', 'launchState must reset to Normal');
    assert.equal(get(enabled), true, 'enabled must reset to true');
    assert.equal(get(waitForExit), false, 'waitForExit must reset to false');
    assert.equal(get(tcpPortProbe), undefined, 'tcpPortProbe must reset to undefined');
    assert.equal(get(probeTimeoutSeconds), 10.0, 'probeTimeoutSeconds must reset to 10.0');

    cleanup();
  });

  test('TC_MODAL_02_AddItem_AfterEditItemPopulatesAndClears', async () => {
    let isOpen = state(false);
    let item = state(null);

    let name = state('');
    let path = state('');

    const cleanup = effect_root(() => {
      user_effect(() => {
        if (get(isOpen)) {
          const it = get(item);
          if (it) {
            set(name, it.name || '');
            set(path, it.path || '');
          } else {
            set(name, '');
            set(path, '');
          }
        }
      });
    });

    await tick();

    // 1. Edit existing item
    set(isOpen, true);
    set(item, { name: 'ExistingApp', path: '/bin/existing' });
    await tick();

    assert.equal(get(name), 'ExistingApp');
    assert.equal(get(path), '/bin/existing');

    // 2. Close modal
    set(isOpen, false);
    await tick();

    // 3. Open Add Item (item set to null)
    set(item, null);
    set(isOpen, true);
    await tick();

    assert.equal(get(name), '', 'name must be cleared when switching to Add Item');
    assert.equal(get(path), '', 'path must be cleared when switching to Add Item');

    cleanup();
  });

  test('TC_MODAL_03_SettingsModal_ResetOnCancel', async () => {
    let isOpen = state(false);
    let settings = state({ startMinimized: false, closeToTray: true });

    let startMinimized = state(false);
    let closeToTray = state(true);

    const cleanup = effect_root(() => {
      user_effect(() => {
        if (get(isOpen) && get(settings)) {
          const s = get(settings);
          set(startMinimized, s.startMinimized ?? false);
          set(closeToTray, s.closeToTray ?? true);
        }
      });
    });

    await tick();

    // 1. Open Settings modal
    set(isOpen, true);
    await tick();

    // 2. User toggles startMinimized to true without saving
    set(startMinimized, true);
    await tick();
    assert.equal(get(startMinimized), true);

    // 3. User cancels/closes modal
    set(isOpen, false);
    await tick();

    // 4. User reopens Settings modal
    set(isOpen, true);
    await tick();

    // Should revert to original saved settings
    assert.equal(get(startMinimized), false, 'startMinimized must revert to saved settings');

    cleanup();
  });
});
