<script lang="ts">
  import type { Profile } from '../types';

  interface Props {
    profiles: Profile[];
    activeProfileName: string;
    onSelectProfile: (name: string) => void;
    onCreateProfile: () => void;
    onDeleteProfile: (name: string) => void;
  }

  let { profiles, activeProfileName, onSelectProfile, onCreateProfile, onDeleteProfile }: Props = $props();
</script>

<div class="flex items-center justify-between px-6 py-2.5 bg-zinc-900/30 border-b border-zinc-800/80 text-xs text-zinc-300">
  <div class="flex items-center gap-3">
    <span class="font-semibold text-zinc-400 uppercase tracking-wider text-[10px]">Launch Profile:</span>
    <div class="flex items-center gap-1.5 bg-zinc-800/90 rounded-lg p-1 border border-zinc-700/60 shadow-inner">
      {#each profiles as p}
        <button
          onclick={() => onSelectProfile(p.name)}
          class="px-3 py-1 rounded-md transition-all font-medium {activeProfileName === p.name ? 'bg-indigo-600 text-white shadow-sm font-semibold' : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-700/50'}"
        >
          {p.name}
          <span class="ml-1 text-[10px] opacity-70">({p.items.length})</span>
        </button>
      {/each}
      
      <button
        onclick={onCreateProfile}
        class="px-2 py-1 rounded-md text-zinc-400 hover:text-indigo-400 hover:bg-zinc-700/50 transition-all font-bold"
        title="Create New Profile"
      >
        +
      </button>
    </div>
  </div>

  <div class="flex items-center gap-2">
    {#if profiles.length > 1}
      <button
        onclick={() => onDeleteProfile(activeProfileName)}
        class="text-zinc-500 hover:text-rose-400 transition-colors text-[11px]"
        title="Delete current profile"
      >
        Delete Profile
      </button>
    {/if}
  </div>
</div>
