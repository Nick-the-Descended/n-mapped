// Tiny app-wide event bus for keyboard-driven actions. Components subscribe
// via on(); shortcuts in App.svelte emit() via window CustomEvents so we
// don't have to thread refs through every component boundary.

export const EVT = {
  focusSearch:  'n-mapped:focus-search',
  runScan:      'n-mapped:run-scan',
  stopScan:     'n-mapped:stop-scan',
  saveFavorite: 'n-mapped:save-favorite',
} as const;

export function emit(name: string) {
  if (typeof window === 'undefined') return;
  window.dispatchEvent(new CustomEvent(name));
}

export function on(name: string, handler: () => void): () => void {
  const fn = handler as EventListener;
  window.addEventListener(name, fn);
  return () => window.removeEventListener(name, fn);
}
