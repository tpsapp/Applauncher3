<script lang="ts">
  import type { LaunchItem } from '../types';
  import QueueItemRow from './QueueItemRow.svelte';

  interface Props {
    items: LaunchItem[];
    itemStatuses: Record<string, string>;
    errorMessages: Record<string, string>;
    onToggle: (id: string) => void;
    onEdit: (item: LaunchItem) => void;
    onDelete: (id: string) => void;
    onMoveUp: (index: number) => void;
    onMoveDown: (index: number) => void;
    onAddItem: () => void;
  }

  let { items, itemStatuses, errorMessages, onToggle, onEdit, onDelete, onMoveUp, onMoveDown, onAddItem }: Props = $props();
</script>

<div class="flex-1 overflow-y-auto px-6 py-4">
  {#if items.length === 0}
    <div class="h-full flex flex-col items-center justify-center text-center p-8 border-2 border-dashed border-zinc-800 rounded-2xl">
      <div class="w-12 h-12 rounded-2xl bg-zinc-900 border border-zinc-800 flex items-center justify-center text-2xl mb-3 text-zinc-500">
        📭
      </div>
      <h3 class="text-sm font-semibold text-zinc-200">No applications in this profile queue</h3>
      <p class="text-xs text-zinc-500 max-w-sm mt-1 mb-4">
        Add applications, binaries, or shell scripts to launch them in sequential order with configurable staggered delays.
      </p>
      <button
        onclick={onAddItem}
        class="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-lg shadow-indigo-600/20 transition-all active:scale-95"
      >
        + Add First Application
      </button>
    </div>
  {:else}
    <div class="flex flex-col gap-2.5">
      {#each items as item, index (item.id)}
        <QueueItemRow
          {item}
          {index}
          total={items.length}
          status={itemStatuses[item.id] || ''}
          errorMessage={errorMessages[item.id]}
          {onToggle}
          {onEdit}
          {onDelete}
          {onMoveUp}
          {onMoveDown}
        />
      {/each}
    </div>
  {/if}
</div>
