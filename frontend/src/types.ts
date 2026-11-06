export type WindowState = 'Normal' | 'Minimized' | 'Hidden';

export interface LaunchItem {
  id: string;
  name: string;
  path: string;
  arguments: string[];
  workingDirectory: string;
  delaySeconds: number;
  launchState: WindowState;
  enabled: boolean;
  waitForExit?: boolean;
  tcpPortProbe?: number;
  probeTimeoutSeconds?: number;
}

export interface Profile {
  name: string;
  items: LaunchItem[];
}

export interface AppSettings {
  startMinimized: boolean;
  closeToTray: boolean;
  confirmOnCancel: boolean;
  theme: 'system' | 'dark' | 'light';
  debugLogging: boolean;
}

export interface LauncherConfig {
  version: number;
  activeProfile: string;
  settings: AppSettings;
  profiles: Profile[];
}

export type QueueState = 'Idle' | 'Running' | 'Paused' | 'Cancelled' | 'Completed';

export interface ProgressEvent {
  currentIndex: number;
  totalItems: number;
  currentItem: LaunchItem;
  itemStatus: string;
  state: QueueState;
  delayRemaining: number;
  errorMessage?: string;
}
