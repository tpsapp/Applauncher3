<script lang="ts">
  interface Props {
    isOpen: boolean;
    existingNames: string[];
    onCreate: (name: string) => void;
    onClose: () => void;
  }

  let { isOpen, existingNames, onCreate, onClose }: Props = $props();

  let profileName = $state('');
  let errorMessage = $state('');

  $effect(() => {
    if (isOpen) {
      profileName = '';
      errorMessage = '';
    }
  });

  function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    const trimmed = profileName.trim();
    if (!trimmed) {
      errorMessage = 'Profile name cannot be empty.';
      return;
    }
    if (existingNames.some(n => n.toLowerCase() === trimmed.toLowerCase())) {
      errorMessage = 'A profile with this name already exists.';
      return;
    }
    onCreate(trimmed);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      onClose();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if isOpen}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-in fade-in duration-150">
    <div class="w-full max-w-sm bg-zinc-900 border border-zinc-800 rounded-2xl shadow-2xl p-6">
      <div class="flex items-center justify-between pb-3 border-b border-zinc-800">
        <div class="flex items-center gap-2">
          <span class="text-base">📁</span>
          <h2 class="text-sm font-bold text-zinc-100">Create Launch Profile</h2>
        </div>
        <button
          onclick={onClose}
          class="text-zinc-500 hover:text-zinc-300 text-lg leading-none"
        >
          ✕
        </button>
      </div>

      <form onsubmit={handleSubmit} class="space-y-4 pt-4 text-xs">
        <div>
          <label for="new-profile-name" class="block font-semibold text-zinc-300 mb-1">
            Profile Name
          </label>
          <input
            id="new-profile-name"
            type="text"
            bind:value={profileName}
            placeholder="e.g. Workstation, Gaming, Kiosk"
            class="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2.5 text-zinc-100 placeholder-zinc-500 focus:outline-none focus:border-indigo-500 font-medium"
          />
          {#if errorMessage}
            <p class="text-[11px] text-rose-400 mt-1.5 flex items-center gap-1">
              <span>⚠️</span> {errorMessage}
            </p>
          {/if}
        </div>

        <p class="text-[11px] text-zinc-500">
          Profiles store independent queues of applications, working directories, and wait intervals.
        </p>

        <div class="flex items-center justify-end gap-2 pt-3 border-t border-zinc-800">
          <button
            type="button"
            onclick={onClose}
            class="px-3.5 py-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-300 font-medium transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            class="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white font-semibold transition-colors shadow-md shadow-indigo-600/20 active:scale-95"
          >
            Create Profile
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
