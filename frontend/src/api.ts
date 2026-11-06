import type { LauncherConfig, ProgressEvent } from './types';

export interface ItemMetadata {
  name: string;
  workingDirectory: string;
}

// Declare Wails runtime globals
declare global {
  interface Window {
    go?: {
      main: {
        App: {
          GetConfig(): Promise<LauncherConfig>;
          SaveConfig(cfg: LauncherConfig): Promise<void>;
          StartQueue(profileName: string): Promise<void>;
          PauseQueue(): Promise<void>;
          ResumeQueue(): Promise<void>;
          CancelQueue(): Promise<void>;
          GetQueueState(): Promise<string>;
          SelectExecutableDialog(): Promise<string>;
          SelectDirectoryDialog(): Promise<string>;
          InferItemMetadata(execPath: string): Promise<ItemMetadata>;
        };
      };
    };
    runtime?: {
      EventsOn(eventName: string, callback: (...args: any[]) => void): () => void;
      EventsEmit(eventName: string, ...args: any[]): void;
    };
  }
}

const mockConfig: LauncherConfig = {
  version: 1,
  activeProfile: 'Workstation',
  settings: {
    startMinimized: false,
    closeToTray: true,
    confirmOnCancel: true,
    theme: 'dark',
    debugLogging: false,
  },
  profiles: [
    {
      name: 'Workstation',
      items: [
        {
          id: 'item-1',
          name: 'Google Chrome',
          path: '/usr/bin/google-chrome-stable',
          arguments: ['--new-window', 'https://github.com'],
          workingDirectory: '',
          delaySeconds: 2.0,
          launchState: 'Normal',
          enabled: true,
        },
        {
          id: 'item-2',
          name: 'VS Code',
          path: '/usr/bin/code',
          arguments: ['.'],
          workingDirectory: '/home/tpsapp/Documents/Projects',
          delaySeconds: 3.5,
          launchState: 'Normal',
          enabled: true,
        },
        {
          id: 'item-3',
          name: 'Terminal Workspace',
          path: '/usr/bin/konsole',
          arguments: ['--workdir', '/home/tpsapp'],
          workingDirectory: '',
          delaySeconds: 1.0,
          launchState: 'Normal',
          enabled: true,
        },
      ],
    },
    {
      name: 'Gaming Mode',
      items: [
        {
          id: 'game-1',
          name: 'Steam',
          path: '/usr/bin/steam',
          arguments: ['-silent'],
          workingDirectory: '',
          delaySeconds: 4.0,
          launchState: 'Minimized',
          enabled: true,
        },
      ],
    },
  ],
};

export const api = {
  async getConfig(): Promise<LauncherConfig> {
    if (window.go?.main?.App) {
      return await window.go.main.App.GetConfig();
    }
    return JSON.parse(JSON.stringify(mockConfig));
  },

  async saveConfig(cfg: LauncherConfig): Promise<void> {
    if (window.go?.main?.App) {
      await window.go.main.App.SaveConfig(cfg);
    } else {
      console.log('[Mock API] Saved config:', cfg);
    }
  },

  async startQueue(profileName: string): Promise<void> {
    if (window.go?.main?.App) {
      await window.go.main.App.StartQueue(profileName);
    } else {
      console.log('[Mock API] Start queue:', profileName);
    }
  },

  async pauseQueue(): Promise<void> {
    if (window.go?.main?.App) {
      await window.go.main.App.PauseQueue();
    } else {
      console.log('[Mock API] Pause queue');
    }
  },

  async resumeQueue(): Promise<void> {
    if (window.go?.main?.App) {
      await window.go.main.App.ResumeQueue();
    } else {
      console.log('[Mock API] Resume queue');
    }
  },

  async cancelQueue(): Promise<void> {
    if (window.go?.main?.App) {
      await window.go.main.App.CancelQueue();
    } else {
      console.log('[Mock API] Cancel queue');
    }
  },

  async selectExecutable(): Promise<string> {
    if (window.go?.main?.App) {
      return await window.go.main.App.SelectExecutableDialog();
    }
    return '/usr/bin/example-app';
  },

  async selectDirectory(): Promise<string> {
    if (window.go?.main?.App) {
      return await window.go.main.App.SelectDirectoryDialog();
    }
    return '/home/tpsapp/Documents';
  },

  async inferMetadata(execPath: string): Promise<ItemMetadata> {
    if (window.go?.main?.App) {
      return await window.go.main.App.InferItemMetadata(execPath);
    }
    const base = execPath.split('/').pop() || '';
    const lastSlash = execPath.lastIndexOf('/');
    const dir = lastSlash > 0 ? execPath.substring(0, lastSlash) : '/';
    return {
      name: base.replace(/\.[^/.]+$/, ''),
      workingDirectory: dir,
    };
  },

  onProgress(callback: (event: ProgressEvent) => void): () => void {
    if (window.runtime?.EventsOn) {
      return window.runtime.EventsOn('queue:progress', callback);
    }
    return () => {};
  },

  onCompleted(callback: (profileName: string) => void): () => void {
    if (window.runtime?.EventsOn) {
      return window.runtime.EventsOn('queue:completed', callback);
    }
    return () => {};
  },
};
