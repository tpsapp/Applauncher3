<script lang="ts">
  import type { AppSettings } from '../types';

  interface Props {
    settings: AppSettings;
    isOpen: boolean;
    onSave: (settings: AppSettings) => void;
    onClose: () => void;
  }

  let { settings, isOpen, onSave, onClose }: Props = $props();

  let startMinimized = $state(false);
  let closeToTray = $state(true);
  let confirmOnCancel = $state(true);
  let theme = $state<'system' | 'dark' | 'light'>('system');
  let debugLogging = $state(false);

  $effect(() => {
    if (isOpen && settings) {
      startMinimized = settings.startMinimized ?? false;
      closeToTray = settings.closeToTray ?? true;
      confirmOnCancel = settings.confirmOnCancel ?? true;
      theme = settings.theme || 'system';
      debugLogging = settings.debugLogging ?? false;
    }
  });

  function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    onSave({
      startMinimized,
      closeToTray,
      confirmOnCancel,
      theme,
      debugLogging,
    });
  }
</script>

{#if isOpen}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-in fade-in duration-200">
    <div class="w-full max-w-md bg-zinc-900 border border-zinc-800 rounded-2xl shadow-2xl p-6">
      <div class="flex items-center justify-between pb-4 border-b border-zinc-800">
        <h2 class="text-base font-bold text-zinc-100 flex items-center gap-2">
          <span>⚙️</span> Preferences & Launcher Settings
        </h2>
        <button onclick={onClose} class="text-zinc-500 hover:text-zinc-300 text-lg leading-none">✕</button>
      </div>

      <form onsubmit={handleSubmit} class="space-y-4 pt-4 text-xs">
        <div class="space-y-3">
          <label class="flex items-center justify-between p-3 rounded-xl bg-zinc-950/60 border border-zinc-800/80 cursor-pointer">
            <div>
              <span class="font-semibold text-zinc-200 block">Close to System Tray</span>
              <span class="text-zinc-500 text-[11px]">Keep launcher running in tray when window is closed (uncheck to exit immediately)</span>
            </div>
            <input type="checkbox" bind:checked={closeToTray} class="w-4 h-4 accent-indigo-500 rounded" />
          </label>

          <label class="flex items-center justify-between p-3 rounded-xl bg-zinc-950/60 border border-zinc-800/80 cursor-pointer">
            <div>
              <span class="font-semibold text-zinc-200 block">Start Minimized</span>
              <span class="text-zinc-500 text-[11px]">Launch silently to system tray on startup</span>
            </div>
            <input type="checkbox" bind:checked={startMinimized} class="w-4 h-4 accent-indigo-500 rounded" />
          </label>

          <label class="flex items-center justify-between p-3 rounded-xl bg-zinc-950/60 border border-zinc-800/80 cursor-pointer">
            <div>
              <span class="font-semibold text-zinc-200 block">Diagnostic Debug Logging</span>
              <span class="text-zinc-500 text-[11px]">Record detailed trace logs to launcher-debug.log</span>
            </div>
            <input type="checkbox" bind:checked={debugLogging} class="w-4 h-4 accent-indigo-500 rounded" />
          </label>
        </div>

        <div>
          <label for="setting-theme" class="block font-semibold text-zinc-300 mb-1">Color Theme</label>
          <select id="setting-theme" bind:value={theme} class="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-zinc-100 focus:outline-none focus:border-indigo-500">
            <option value="system">Follow System Theme</option>
            <option value="dark">Dark Theme</option>
            <option value="light">Light Theme</option>
          </select>
        </div>

        <div class="flex items-center justify-end gap-2 pt-4 border-t border-zinc-800">
          <button type="button" onclick={onClose} class="px-4 py-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-300 font-medium transition-colors">
            Cancel
          </button>
          <button type="submit" class="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white font-semibold transition-colors shadow-md shadow-indigo-600/20">
            Save Preferences
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
