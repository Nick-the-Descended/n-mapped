// Theme manager. Stores the user's preference under localStorage["n-mapped:theme"]
// = "light" | "dark" | "system". Applies it by setting data-theme on the
// document element; CSS in app.css uses [data-theme="..."] selectors.

const KEY = 'n-mapped:theme';
export type Theme = 'light' | 'dark' | 'system';

export function loadTheme(): Theme {
  try {
    const v = localStorage.getItem(KEY);
    if (v === 'light' || v === 'dark' || v === 'system') return v;
  } catch { /* private mode */ }
  return 'system';
}

export function saveTheme(t: Theme) {
  try { localStorage.setItem(KEY, t); } catch {}
  applyTheme(t);
}

export function applyTheme(t: Theme) {
  const el = document.documentElement;
  if (t === 'system') el.removeAttribute('data-theme');
  else el.setAttribute('data-theme', t);
}
