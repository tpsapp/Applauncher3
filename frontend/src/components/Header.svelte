<script lang="ts">
  interface Props {
    onOpenSettings: () => void;
    onAddItem: () => void;
  }
  let { onOpenSettings, onAddItem }: Props = $props();
</script>

<header class="flex items-center justify-between px-6 py-4 border-b border-zinc-800 bg-zinc-900/60 backdrop-blur-md select-none">
  <div class="flex items-center gap-3">
    <div class="w-9 h-9 rounded-xl bg-gradient-to-tr from-indigo-600 via-indigo-500 to-cyan-400 flex items-center justify-center shadow-lg shadow-indigo-500/20 text-white font-black text-lg">
      🚀
    </div>
    <div>
      <h1 class="text-lg font-bold text-zinc-100 tracking-tight flex items-center gap-2">
        AppLauncher<span class="text-indigo-400 text-sm font-semibold px-1.5 py-0.5 rounded bg-indigo-500/10 border border-indigo-500/20">v3</span>
      </h1>
      <p class="text-xs text-zinc-400">Sequential & Staggered Process Automator</p>
    </div>
  </div>

  <div class="flex items-center gap-2">
    <button
      onclick={onAddItem}
      class="px-3.5 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium flex items-center gap-1.5 shadow-md shadow-indigo-600/20 transition-all active:scale-95"
    >
      <span class="text-base leading-none font-bold">+</span>
      <span>Add Item</span>
    </button>

    <button
      onclick={onOpenSettings}
      class="p-2 rounded-lg bg-zinc-800/80 hover:bg-zinc-700 text-zinc-300 hover:text-white transition-all text-xs font-medium"
      title="Preferences & Settings"
    >
      ⚙️
    </button>
  </div>
</header>
