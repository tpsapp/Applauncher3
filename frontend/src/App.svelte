<script lang="ts">
  import { onMount } from 'svelte';
  import type { LauncherConfig, Profile, LaunchItem, QueueState, ProgressEvent, AppSettings } from './types';
  import { api } from './api';

  import Header from './components/Header.svelte';
  import ProfileBar from './components/ProfileBar.svelte';
  import QueueTable from './components/QueueTable.svelte';
  import ItemModal from './components/ItemModal.svelte';
  import SettingsModal from './components/SettingsModal.svelte';
  import ControlBar from './components/ControlBar.svelte';

  let config = $state<LauncherConfig | null>(null);
  let activeProfileName = $state('Default');
  let queueState = $state<QueueState>('Idle');
  let currentIndex = $state(0);
  let delayRemaining = $state(0);
  let itemStatuses = $state<Record<string, string>>({});
  let errorMessages = $state<Record<string, string>>({});

  let isItemModalOpen = $state(false);
  let editingItem = $state<LaunchItem | null>(null);
  let isSettingsOpen = $state(false);

  let activeProfile = $derived<Profile | null>(() => {
    if (!config) return null;
    return config.profiles.find(p => p.name === activeProfileName) || config.profiles[0] || null;
  });

  onMount(async () => {
    try {
      config = await api.getConfig();
      if (config) {
        activeProfileName = config.activeProfile || config.profiles[0]?.name || 'Default';
      }
    } catch (e) {
      console.error('Failed to load initial config:', e);
    }

    // Register Wails backend event listeners
    const unbindProgress = api.onProgress((evt: ProgressEvent) => {
      currentIndex = evt.currentIndex;
      queueState = evt.state;
      delayRemaining = evt.delayRemaining || 0;
      if (evt.currentItem?.id) {
        itemStatuses[evt.currentItem.id] = evt.itemStatus;
        if (evt.errorMessage) {
          errorMessages[evt.currentItem.id] = evt.errorMessage;
        }
      }
    });

    const unbindCompleted = api.onCompleted(() => {
      queueState = 'Completed';
      delayRemaining = 0;
    });

    return () => {
      unbindProgress();
      unbindCompleted();
    };
  });

  async function persistConfig() {
    if (!config) return;
    config.activeProfile = activeProfileName;
    try {
      await api.saveConfig(config);
    } catch (e) {
      console.error('Failed to save config:', e);
    }
  }

  function handleSelectProfile(name: string) {
    activeProfileName = name;
    itemStatuses = {};
    errorMessages = {};
    currentIndex = 0;
    queueState = 'Idle';
    persistConfig();
  }

  function handleCreateProfile() {
    if (!config) return;
    const name = prompt('Enter new profile name:');
    if (!name || !name.trim()) return;
    const trimmed = name.trim();
    if (config.profiles.some(p => p.name === trimmed)) {
      alert('A profile with that name already exists.');
      return;
    }
    config.profiles.push({
      name: trimmed,
      items: [],
    });
    activeProfileName = trimmed;
    persistConfig();
  }

  function handleDeleteProfile(name: string) {
    if (!config || config.profiles.length <= 1) return;
    if (!confirm(`Are you sure you want to delete profile "${name}"?`)) return;
    config.profiles = config.profiles.filter(p => p.name !== name);
    activeProfileName = config.profiles[0].name;
    persistConfig();
  }

  function handleToggleItem(id: string) {
    const prof = activeProfile();
    if (!prof) return;
    const item = prof.items.find(i => i.id === id);
    if (item) {
      item.enabled = !item.enabled;
      persistConfig();
    }
  }

  function handleMoveUp(index: number) {
    const prof = activeProfile();
    if (!prof || index <= 0) return;
    const temp = prof.items[index];
    prof.items[index] = prof.items[index - 1];
    prof.items[index - 1] = temp;
    persistConfig();
  }

  function handleMoveDown(index: number) {
    const prof = activeProfile();
    if (!prof || index >= prof.items.length - 1) return;
    const temp = prof.items[index];
    prof.items[index] = prof.items[index + 1];
    prof.items[index + 1] = temp;
    persistConfig();
  }

  function handleDeleteItem(id: string) {
    const prof = activeProfile();
    if (!prof) return;
    prof.items = prof.items.filter(i => i.id !== id);
    persistConfig();
  }

  function handleOpenAddItem() {
    editingItem = null;
    isItemModalOpen = true;
  }

  function handleOpenEditItem(item: LaunchItem) {
    editingItem = JSON.parse(JSON.stringify(item));
    isItemModalOpen = true;
  }

  function handleSaveItem(saved: LaunchItem) {
    const prof = activeProfile();
    if (!prof) return;

    if (saved.id) {
      const idx = prof.items.findIndex(i => i.id === saved.id);
      if (idx >= 0) {
        prof.items[idx] = saved;
      } else {
        prof.items.push(saved);
      }
    } else {
      saved.id = Math.random().toString(36).substring(2, 10);
      prof.items.push(saved);
    }

    isItemModalOpen = false;
    persistConfig();
  }

  async function handleStartQueue() {
    itemStatuses = {};
    errorMessages = {};
    currentIndex = 0;
    queueState = 'Running';
    try {
      await api.startQueue(activeProfileName);
    } catch (e) {
      console.error('Error starting queue:', e);
      queueState = 'Idle';
    }
  }

  async function handlePauseQueue() {
    try {
      await api.pauseQueue();
      queueState = 'Paused';
    } catch (e) {
      console.error(e);
    }
  }

  async function handleResumeQueue() {
    try {
      await api.resumeQueue();
      queueState = 'Running';
    } catch (e) {
      console.error(e);
    }
  }

  async function handleCancelQueue() {
    try {
      await api.cancelQueue();
      queueState = 'Cancelled';
    } catch (e) {
      console.error(e);
    }
  }

  function handleSaveSettings(newSettings: AppSettings) {
    if (!config) return;
    config.settings = newSettings;
    isSettingsOpen = false;
    persistConfig();
  }
</script>

<main class="h-screen flex flex-col bg-zinc-950 text-zinc-100 antialiased overflow-hidden select-none">
  <Header
    onOpenSettings={() => (isSettingsOpen = true)}
    onAddItem={handleOpenAddItem}
  />

  {#if config}
    <ProfileBar
      profiles={config.profiles}
      {activeProfileName}
      onSelectProfile={handleSelectProfile}
      onCreateProfile={handleCreateProfile}
      onDeleteProfile={handleDeleteProfile}
    />

    <QueueTable
      items={activeProfile()?.items || []}
      {itemStatuses}
      {errorMessages}
      onToggle={handleToggleItem}
      onEdit={handleOpenEditItem}
      onDelete={handleDeleteItem}
      onMoveUp={handleMoveUp}
      onMoveDown={handleMoveDown}
      onAddItem={handleOpenAddItem}
    />

    <ControlBar
      state={queueState}
      {currentIndex}
      totalItems={activeProfile()?.items?.length || 0}
      {delayRemaining}
      onStart={handleStartQueue}
      onPause={handlePauseQueue}
      onResume={handleResumeQueue}
      onCancel={handleCancelQueue}
    />
  {/if}

  <ItemModal
    item={editingItem}
    isOpen={isItemModalOpen}
    onSave={handleSaveItem}
    onClose={() => (isItemModalOpen = false)}
  />

  {#if config}
    <SettingsModal
      settings={config.settings}
      isOpen={isSettingsOpen}
      onSave={handleSaveSettings}
      onClose={() => (isSettingsOpen = false)}
    />
  {/if}
</main>
