// Builder state — what flags the user has selected and what target they typed.
// Kept in module scope as a plain reactive object that components import.

import type { Flag } from './types';

export interface BuilderState {
  selectedFlagIds: Set<string>;
  flagValues: Record<string, string>;
  targets: string;
}

// Plain singleton; components mutate fields directly. We deliberately avoid
// Svelte 5 runes in this module so the store can be shared across components
// without prop-drilling. Components should subscribe via $state mirrors when
// they need to react to changes.
export const builder: BuilderState = {
  selectedFlagIds: new Set<string>(),
  flagValues: {},
  targets: '',
};

// Build the canonical command preview from the current builder state.
export function previewCommand(allFlags: Flag[], targets: string, selected: Set<string>, values: Record<string, string>): string {
  const parts: string[] = ['nmap'];
  const sorted = [...selected].sort();
  for (const id of sorted) {
    const flag = allFlags.find((f) => f.id === id);
    if (!flag) continue;
    const tok = flag.short || flag.long;
    if (!tok) continue;
    parts.push(tok);
    if (flag.value_type && flag.value_type !== 'boolean') {
      const v = values[id];
      if (v) parts.push(v);
      else parts.push('<value>');
    }
  }
  parts.push('-oX', '-');
  parts.push('--stats-every', '2s');
  const tlist = targets
    .split(/[\s,]+/)
    .map((t) => t.trim())
    .filter(Boolean);
  if (tlist.length > 0) {
    parts.push('--', ...tlist);
  } else {
    parts.push('--', '<target>');
  }
  return parts.join(' ');
}

// Plain-English summary from selected flag short_descriptions.
export function previewSummary(allFlags: Flag[], selected: Set<string>): string {
  if (selected.size === 0) return 'Pick one or more options below to build a scan.';
  const lines: string[] = [];
  for (const id of selected) {
    const f = allFlags.find((x) => x.id === id);
    if (f) lines.push('• ' + f.short_description);
  }
  return lines.join('\n');
}
