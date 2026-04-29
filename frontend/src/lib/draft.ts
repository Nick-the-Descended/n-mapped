// localStorage-backed draft of the builder state. Saves on every change
// (debounced) so a tab refresh doesn't lose the user's work.

const KEY = 'n-mapped:draft';

export interface Draft {
  targets: string;
  selectedFlagIDs: string[];
  flagValues: Record<string, string>;
  selectedScriptIDs: string[];
  scriptArgs: Record<string, string>;
}

export function loadDraft(): Draft | null {
  try {
    const v = localStorage.getItem(KEY);
    if (!v) return null;
    const parsed = JSON.parse(v) as unknown;
    if (!parsed || typeof parsed !== 'object') return null;
    const d = parsed as Partial<Draft>;
    return {
      targets: d.targets ?? '',
      selectedFlagIDs: Array.isArray(d.selectedFlagIDs) ? d.selectedFlagIDs : [],
      flagValues: d.flagValues && typeof d.flagValues === 'object' ? d.flagValues : {},
      selectedScriptIDs: Array.isArray(d.selectedScriptIDs) ? d.selectedScriptIDs : [],
      scriptArgs: d.scriptArgs && typeof d.scriptArgs === 'object' ? d.scriptArgs : {},
    };
  } catch {
    return null;
  }
}

export function saveDraft(d: Draft) {
  try { localStorage.setItem(KEY, JSON.stringify(d)); } catch {}
}

export function clearDraft() {
  try { localStorage.removeItem(KEY); } catch {}
}
