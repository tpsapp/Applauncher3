export type ThemeMode = 'system' | 'dark' | 'light';

export function getResolvedTheme(mode: ThemeMode): 'dark' | 'light' {
  if (mode === 'light') return 'light';
  if (mode === 'dark') return 'dark';
  if (typeof window !== 'undefined' && window.matchMedia) {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }
  return 'dark';
}

export function applyTheme(mode: ThemeMode) {
  if (typeof document === 'undefined') return;
  const resolved = getResolvedTheme(mode);
  const root = document.documentElement;
  if (resolved === 'dark') {
    root.classList.remove('light');
    root.classList.add('dark');
  } else {
    root.classList.remove('dark');
    root.classList.add('light');
  }
}

export function setupSystemThemeListener(getTheme: () => ThemeMode): () => void {
  if (typeof window === 'undefined' || !window.matchMedia) {
    return () => {};
  }
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
  const handler = () => {
    if (getTheme() === 'system') {
      applyTheme('system');
    }
  };
  mediaQuery.addEventListener('change', handler);
  return () => mediaQuery.removeEventListener('change', handler);
}
