<script lang="ts">
  import type { QueueState } from '../types';

  interface Props {
    state: QueueState;
    currentIndex: number;
    totalItems: number;
    delayRemaining: number;
    onStart: () => void;
    onPause: () => void;
    onResume: () => void;
    onCancel: () => void;
  }

  let { state, currentIndex, totalItems, delayRemaining, onStart, onPause, onResume, onCancel }: Props = $props();

  let progressPercent = $derived(() => {
    if (totalItems === 0) return 0;
    if (state === 'Completed') return 100;
    const basePercent = (currentIndex / totalItems) * 100;
    return Math.min(100, Math.max(0, basePercent));
  });
</script>

<footer class="border-t border-zinc-800 bg-zinc-900/80 backdrop-blur-md px-6 py-4 select-none">
  <!-- Real-Time Progress Bar -->
  <div class="mb-3">
    <div class="flex items-center justify-between text-xs mb-1.5 font-mono">
      <div class="flex items-center gap-2">
        <span class="font-bold text-zinc-300">
          {#if state === 'Running'}
            <span class="inline-block w-2 h-2 rounded-full bg-blue-500 animate-pulse mr-1"></span>
            Launching Step {currentIndex + 1} of {totalItems}
          {:else if state === 'Paused'}
            <span class="inline-block w-2 h-2 rounded-full bg-amber-500 mr-1"></span>
            Paused at Step {currentIndex + 1} of {totalItems}
          {:else if state === 'Completed'}
            <span class="inline-block w-2 h-2 rounded-full bg-emerald-500 mr-1"></span>
            Execution Complete ({totalItems} items)
          {:else if state === 'Cancelled'}
            <span class="inline-block w-2 h-2 rounded-full bg-rose-500 mr-1"></span>
            Cancelled at Step {currentIndex + 1}
          {:else}
            Ready ({totalItems} items configured)
          {/if}
        </span>
      </div>

      {#if delayRemaining > 0 && state === 'Running'}
        <span class="text-indigo-400 font-semibold">
          ⏳ Delay remaining: {delayRemaining.toFixed(1)}s
        </span>
      {:else}
        <span class="text-zinc-500">
          {Math.round(progressPercent())}%
        </span>
      {/if}
    </div>

    <!-- Progress track -->
    <div class="w-full h-2 rounded-full bg-zinc-800 overflow-hidden">
      <div
        class="h-full bg-gradient-to-r from-indigo-500 to-cyan-400 transition-all duration-300 ease-out"
        style="width: {progressPercent()}%;"
      ></div>
    </div>
  </div>

  <!-- Interactive Controls Bar -->
  <div class="flex items-center justify-between pt-1">
    <div class="flex items-center gap-2">
      {#if state === 'Idle' || state === 'Completed' || state === 'Cancelled'}
        <button
          onclick={onStart}
          disabled={totalItems === 0}
          class="px-5 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 disabled:opacity-30 disabled:hover:bg-emerald-600 text-white text-xs font-bold flex items-center gap-2 shadow-lg shadow-emerald-600/20 transition-all active:scale-95"
        >
          <span>▶</span>
          <span>Start Execution</span>
        </button>
      {/if}

      {#if state === 'Running'}
        <button
          onclick={onPause}
          class="px-5 py-2 rounded-xl bg-amber-600 hover:bg-amber-500 text-white text-xs font-bold flex items-center gap-2 shadow-lg shadow-amber-600/20 transition-all active:scale-95"
        >
          <span>⏸</span>
          <span>Pause Queue</span>
        </button>
      {/if}

      {#if state === 'Paused'}
        <button
          onclick={onResume}
          class="px-5 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold flex items-center gap-2 shadow-lg shadow-indigo-600/20 transition-all active:scale-95"
        >
          <span>▶</span>
          <span>Resume Queue</span>
        </button>
      {/if}

      {#if state === 'Running' || state === 'Paused'}
        <button
          onclick={onCancel}
          class="px-4 py-2 rounded-xl bg-zinc-800 hover:bg-rose-900/60 text-zinc-300 hover:text-rose-200 text-xs font-semibold flex items-center gap-1.5 transition-all active:scale-95 border border-zinc-700/60"
        >
          <span>⏹</span>
          <span>Cancel</span>
        </button>
      {/if}
    </div>

    <div class="text-right text-[11px] text-zinc-500">
      <span>Mode: Detached Child Processes</span>
    </div>
  </div>
</footer>
