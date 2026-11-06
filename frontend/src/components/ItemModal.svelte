<script lang="ts">
  import type { LaunchItem, WindowState } from '../types';
  import { api } from '../api';

  interface Props {
    item: LaunchItem | null;
    isOpen: boolean;
    onSave: (item: LaunchItem) => void;
    onClose: () => void;
  }

  let { item, isOpen, onSave, onClose }: Props = $props();

  let name = $state('');
  let path = $state('');
  let args = $state('');
  let workingDirectory = $state('');
  let delaySeconds = $state(2.0);
  let launchState = $state<WindowState>('Normal');
  let enabled = $state(true);

  $effect(() => {
    if (item) {
      name = item.name || '';
      path = item.path || '';
      args = (item.arguments || []).join(' ');
      workingDirectory = item.workingDirectory || '';
      delaySeconds = item.delaySeconds ?? 2.0;
      launchState = item.launchState || 'Normal';
      enabled = item.enabled ?? true;
    } else {
      name = '';
      path = '';
      args = '';
      workingDirectory = '';
      delaySeconds = 2.0;
      launchState = 'Normal';
      enabled = true;
    }
  });

  async function handleBrowseFile() {
    try {
      const selected = await api.selectExecutable();
      if (selected) {
        path = selected;
        const [inferredName, inferredDir] = await api.inferMetadata(selected);
        if (!name) name = inferredName;
        if (!workingDirectory) workingDirectory = inferredDir;
      }
    } catch (e) {
      console.error('File dialog error:', e);
    }
  }

  async function handleBrowseDir() {
    try {
      const selected = await api.selectDirectory();
      if (selected) {
        workingDirectory = selected;
      }
    } catch (e) {
      console.error('Directory dialog error:', e);
    }
  }

  function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    if (!path.trim()) return;

    const parsedArgs = args.trim()
      ? args.match(/(?:[^\s"]+|"[^"]*")+/g)?.map(a => a.replace(/^"|"$/g, '')) || []
      : [];

    const result: LaunchItem = {
      id: item?.id || '',
      name: name.trim() || path.split('/').pop()?.replace(/\.[^/.]+$/, '') || 'Application',
      path: path.trim(),
      arguments: parsedArgs,
      workingDirectory: workingDirectory.trim(),
      delaySeconds: Math.max(0, parseFloat(delaySeconds.toString()) || 0),
      launchState,
      enabled,
    };

    onSave(result);
  }
</script>

{#if isOpen}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-in fade-in duration-200">
    <div class="w-full max-w-lg bg-zinc-900 border border-zinc-800 rounded-2xl shadow-2xl p-6 overflow-hidden">
      <div class="flex items-center justify-between pb-4 border-b border-zinc-800">
        <h2 class="text-base font-bold text-zinc-100">
          {item ? 'Edit Application' : 'Add Application'}
        </h2>
        <button
          onclick={onClose}
          class="text-zinc-500 hover:text-zinc-300 text-lg leading-none"
        >
          ✕
        </button>
      </div>

      <form onsubmit={handleSubmit} class="space-y-4 pt-4 text-xs">
        <!-- Executable Path & Browse -->
        <div>
          <label for="item-path" class="block font-semibold text-zinc-300 mb-1">Executable / Target Path <span class="text-indigo-400">*</span></label>
          <div class="flex gap-2">
            <input
              id="item-path"
              type="text"
              bind:value={path}
              required
              placeholder="/usr/bin/google-chrome or C:\Program Files\..."
              class="flex-1 bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-zinc-100 font-mono focus:outline-none focus:border-indigo-500"
            />
            <button
              type="button"
              onclick={handleBrowseFile}
              class="px-3 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-200 rounded-lg font-medium transition-colors"
            >
              Browse...
            </button>
          </div>
        </div>

        <!-- Friendly Display Name -->
        <div>
          <label for="item-name" class="block font-semibold text-zinc-300 mb-1">Display Label</label>
          <input
            id="item-name"
            type="text"
            bind:value={name}
            placeholder="e.g. Web Browser"
            class="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-zinc-100 focus:outline-none focus:border-indigo-500"
          />
        </div>

        <!-- Command-Line Arguments -->
        <div>
          <label for="item-args" class="block font-semibold text-zinc-300 mb-1">Command-Line Arguments</label>
          <input
            id="item-args"
            type="text"
            bind:value={args}
            placeholder="--new-window https://example.com"
            class="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-zinc-100 font-mono focus:outline-none focus:border-indigo-500"
          />
          <span class="text-[10px] text-zinc-500 mt-0.5 block">Space-separated arguments. Enclose values with spaces in quotes.</span>
        </div>

        <!-- Working Directory -->
        <div>
          <label for="item-workdir" class="block font-semibold text-zinc-300 mb-1">Working Directory</label>
          <div class="flex gap-2">
            <input
              id="item-workdir"
              type="text"
              bind:value={workingDirectory}
              placeholder="Leave empty to inherit default directory"
              class="flex-1 bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-zinc-100 font-mono focus:outline-none focus:border-indigo-500"
            />
            <button
              type="button"
              onclick={handleBrowseDir}
              class="px-3 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-200 rounded-lg font-medium transition-colors"
            >
              Browse...
            </button>
          </div>
        </div>

        <!-- Delay Interval & Window State -->
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label for="item-delay" class="block font-semibold text-zinc-300 mb-1">Wait Delay: <span class="text-indigo-400 font-mono">{delaySeconds}s</span></label>
            <input
              id="item-delay"
              type="range"
              min="0"
              max="30"
              step="0.5"
              bind:value={delaySeconds}
              class="w-full accent-indigo-500"
            />
          </div>

          <div>
            <label for="item-launchstate" class="block font-semibold text-zinc-300 mb-1">Launch Window State</label>
            <select
              id="item-launchstate"
              bind:value={launchState}
              class="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-zinc-100 focus:outline-none focus:border-indigo-500"
            >
              <option value="Normal">Normal</option>
              <option value="Minimized">Minimized</option>
              <option value="Hidden">Hidden</option>
            </select>
          </div>
        </div>

        <!-- Modal Footer Actions -->
        <div class="flex items-center justify-end gap-2 pt-4 border-t border-zinc-800">
          <button
            type="button"
            onclick={onClose}
            class="px-4 py-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-300 font-medium transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            class="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white font-semibold transition-colors shadow-md shadow-indigo-600/20"
          >
            Save Item
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
