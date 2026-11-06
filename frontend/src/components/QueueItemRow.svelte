<script lang="ts">
  import type { LaunchItem } from '../types';

  interface Props {
    item: LaunchItem;
    index: number;
    total: number;
    status: string;
    errorMessage?: string;
    onToggle: (id: string) => void;
    onEdit: (item: LaunchItem) => void;
    onDelete: (id: string) => void;
    onMoveUp: (index: number) => void;
    onMoveDown: (index: number) => void;
  }

  let { item, index, total, status, errorMessage, onToggle, onEdit, onDelete, onMoveUp, onMoveDown }: Props = $props();

  let statusBadgeClass = $derived(() => {
    switch (status) {
      case 'Running':
        return 'bg-blue-500/20 text-blue-400 border-blue-500/40 animate-pulse';
      case 'Delaying':
        return 'bg-amber-500/20 text-amber-400 border-amber-500/40';
      case 'Success':
        return 'bg-emerald-500/20 text-emerald-400 border-emerald-500/40';
      case 'Failed':
        return 'bg-rose-500/20 text-rose-400 border-rose-500/40';
      case 'Skipped':
        return 'bg-zinc-800 text-zinc-500 border-zinc-700';
      case 'Cancelled':
        return 'bg-zinc-800 text-amber-500 border-zinc-700';
      default:
        return 'bg-zinc-800 text-zinc-400 border-zinc-700';
    }
  });
</script>

<div class="group flex items-center justify-between px-4 py-3 bg-zinc-900/40 hover:bg-zinc-900 border border-zinc-800/80 rounded-xl transition-all duration-150 {item.enabled ? '' : 'opacity-50'}">
  <!-- Left Column: Step & Info -->
  <div class="flex items-center gap-3.5 flex-1 min-w-0 pr-4">
    <!-- Reorder Arrows -->
    <div class="flex flex-col gap-0.5 text-zinc-500">
      <button
        onclick={() => onMoveUp(index)}
        disabled={index === 0}
        class="hover:text-indigo-400 disabled:opacity-20 disabled:hover:text-zinc-500 text-[10px] leading-none p-0.5"
        title="Move Up"
      >
        ▲
      </button>
      <button
        onclick={() => onMoveDown(index)}
        disabled={index === total - 1}
        class="hover:text-indigo-400 disabled:opacity-20 disabled:hover:text-zinc-500 text-[10px] leading-none p-0.5"
        title="Move Down"
      >
        ▼
      </button>
    </div>

    <!-- Step number badge -->
    <span class="w-6 h-6 rounded-md bg-zinc-800 border border-zinc-700 text-zinc-400 text-xs font-mono font-bold flex items-center justify-center shrink-0">
      {index + 1}
    </span>

    <!-- Item Metadata -->
    <div class="min-w-0 flex-1">
      <div class="flex items-center gap-2">
        <h3 class="text-sm font-semibold text-zinc-100 truncate">{item.name || 'Unnamed Application'}</h3>
        <span class="px-2 py-0.5 rounded-full text-[10px] font-mono border {statusBadgeClass()}">
          {status || (item.enabled ? 'Pending' : 'Disabled')}
        </span>
        {#if item.waitForExit}
          <span class="px-1.5 py-0.5 rounded text-[10px] bg-amber-500/10 text-amber-400 border border-amber-500/20 font-mono" title="Waits for process exit before continuing">
            Blocking
          </span>
        {/if}
        {#if item.tcpPortProbe}
          <span class="px-1.5 py-0.5 rounded text-[10px] bg-cyan-500/10 text-cyan-400 border border-cyan-500/20 font-mono" title="Probes TCP port {item.tcpPortProbe}">
            TCP :{item.tcpPortProbe}
          </span>
        {/if}
        {#if errorMessage}
          <span class="text-xs text-rose-400 font-mono truncate" title={errorMessage}>
            ⚠️ {errorMessage}
          </span>
        {/if}
      </div>
      <p class="text-xs text-zinc-400 font-mono truncate mt-0.5">
        {item.path}
        {#if item.arguments && item.arguments.length > 0}
          <span class="text-zinc-500"> {item.arguments.join(' ')}</span>
        {/if}
      </p>
    </div>
  </div>

  <!-- Right Column: Settings & Actions -->
  <div class="flex items-center gap-3 shrink-0">
    <!-- Delay indicator -->
    <div class="text-right">
      <span class="text-xs font-mono font-semibold text-indigo-400">{item.delaySeconds.toFixed(1)}s</span>
      <span class="text-[10px] text-zinc-500 block">wait delay</span>
    </div>

    <!-- Window state tag -->
    <span class="text-[10px] px-2 py-1 rounded bg-zinc-800 text-zinc-400 border border-zinc-700/60 font-mono">
      {item.launchState}
    </span>

    <!-- Enable / Disable Switch -->
    <button
      onclick={() => onToggle(item.id)}
      class="w-8 h-4 rounded-full transition-colors relative {item.enabled ? 'bg-indigo-600' : 'bg-zinc-700'}"
      title={item.enabled ? 'Click to disable' : 'Click to enable'}
    >
      <span class="absolute top-0.5 bottom-0.5 w-3 rounded-full bg-white transition-all {item.enabled ? 'right-0.5' : 'left-0.5'}"></span>
    </button>

    <!-- Action Buttons -->
    <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
      <button
        onclick={() => onEdit(item)}
        class="p-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-300 hover:text-white transition-all text-xs"
        title="Edit item configuration"
      >
        ✏️
      </button>
      <button
        onclick={() => onDelete(item.id)}
        class="p-1.5 rounded-lg bg-zinc-800 hover:bg-rose-900/60 text-zinc-400 hover:text-rose-300 transition-all text-xs"
        title="Delete item"
      >
        🗑️
      </button>
    </div>
  </div>
</div>
