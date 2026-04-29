<script lang="ts">
  import NotifyToggle from './NotifyToggle.svelte';
  import type { NmapInfo, PrivilegeState } from '../lib/types';

  let { nmap, privilege, onOpenSettings }: {
    nmap: NmapInfo | null;
    privilege: PrivilegeState | null;
    onOpenSettings: () => void;
  } = $props();

  let nmapPill = $derived.by(() => {
    if (!nmap) return { text: 'detecting nmap…', cls: '' };
    if (!nmap.ok) return { text: 'nmap not found', cls: 'danger' };
    return { text: `nmap ${nmap.version}`, cls: 'ok' };
  });

  let privPill = $derived.by(() => {
    if (!privilege) return { text: 'detecting privilege…', cls: '' };
    switch (privilege.mode) {
      case 'root':       return { text: 'running as root', cls: 'warn' };
      case 'capability': return { text: 'capability mode', cls: 'accent' };
      default:           return { text: 'user mode (raw scans disabled)', cls: '' };
    }
  });
</script>

<header>
  <div class="brand">n-mapped</div>
  <div class="pills">
    <NotifyToggle />
    <span class="pill {nmapPill.cls}" title={nmap?.path ?? ''}>{nmapPill.text}</span>
    <span class="pill {privPill.cls}" title={privilege?.reason ?? ''}>{privPill.text}</span>
    <button class="settings" onclick={onOpenSettings} title="Settings (theme, shortcuts, etc.)" aria-label="Open settings">⚙</button>
  </div>
</header>

<style>
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.85rem 1.25rem;
    border-bottom: 1px solid var(--border);
    background: var(--bg-elev);
  }
  .brand {
    font-weight: 700;
    font-size: 1.1rem;
    letter-spacing: 0.02em;
  }
  .pills {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }
  .settings {
    background: transparent;
    border: 1px solid transparent;
    padding: 0.2rem 0.5rem;
    border-radius: var(--radius);
    color: var(--text-dim);
    font-size: 1rem;
    cursor: pointer;
  }
  .settings:hover { color: var(--text); border-color: var(--border-strong); }
</style>
