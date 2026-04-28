<script lang="ts">
  import { onMount } from 'svelte';
  import Header from './components/Header.svelte';
  import TargetInput from './components/TargetInput.svelte';
  import FlagPicker from './components/FlagPicker.svelte';
  import CommandPreview from './components/CommandPreview.svelte';
  import ScanRunner from './components/ScanRunner.svelte';
  import { api } from './lib/api';
  import type { Catalog, NmapInfo, PrivilegeState, ScanRequest } from './lib/types';
  import { previewCommand, previewSummary } from './lib/store';

  let catalog = $state<Catalog | null>(null);
  let nmap = $state<NmapInfo | null>(null);
  let privilege = $state<PrivilegeState | null>(null);
  let bootError = $state<string | null>(null);

  let targets = $state('');
  let selected = $state<Set<string>>(new Set());
  let values = $state<Record<string, string>>({});

  onMount(async () => {
    try {
      const [cat, ver, priv] = await Promise.all([
        api.catalog(), api.version(), api.privilege(),
      ]);
      catalog = cat;
      nmap = ver;
      privilege = priv;
    } catch (e) {
      bootError = (e as Error).message;
    }
  });

  let command = $derived(
    catalog ? previewCommand(catalog.flags, targets, selected, values) : 'nmap …',
  );
  let summary = $derived(
    catalog ? previewSummary(catalog.flags, selected) : 'loading catalog…',
  );

  let request = $derived<ScanRequest>({
    targets: targets.split(/[\s,]+/).map((t) => t.trim()).filter(Boolean),
    flag_ids: [...selected],
    flag_values: { ...values },
  });
</script>

<Header {nmap} {privilege} />

<main>
  {#if bootError}
    <div class="boot-error">
      Failed to load backend: <code>{bootError}</code>.
      Is <code>n-mapped</code> running? Try <code>./n-mapped --no-browser</code> in a terminal.
    </div>
  {:else if !catalog || !privilege}
    <div class="loading">Loading…</div>
  {:else}
    {#if nmap && !nmap.ok}
      <div class="install-warning">
        <strong>nmap not detected.</strong>
        Install it before running scans: <code>sudo apt install nmap</code>,
        <code>brew install nmap</code>, or grab it from <a href="https://nmap.org/download.html">nmap.org</a>.
      </div>
    {/if}
    {#if privilege.mode === 'user'}
      <div class="priv-hint">
        Some scan types (<code>-sS</code>, <code>-sU</code>, <code>-O</code>, <code>-A</code>)
        need raw sockets. Quit and relaunch with <code>sudo n-mapped</code> to enable them.
      </div>
    {/if}

    <TargetInput bind:value={targets} />
    <FlagPicker {catalog} {privilege} bind:selected bind:values />
    <CommandPreview {command} {summary} />
    <ScanRunner {request} />
  {/if}
</main>

<footer>
  <span>Phase 1 preview &middot; backend on {nmap?.path ?? 'PATH'}</span>
  <a href="https://nmap.org/book/" target="_blank" rel="noopener noreferrer">Nmap reference guide</a>
</footer>

<style>
  main {
    max-width: 880px;
    margin: 0 auto;
    padding: 1.25rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  .boot-error, .install-warning, .priv-hint {
    padding: 0.75rem 1rem;
    border-radius: var(--radius);
    border: 1px solid var(--border);
  }
  .boot-error {
    background: rgba(220, 38, 38, 0.08);
    border-color: rgba(220, 38, 38, 0.2);
    color: var(--danger);
  }
  .install-warning {
    background: rgba(217, 119, 6, 0.08);
    border-color: rgba(217, 119, 6, 0.2);
    color: var(--warn);
  }
  .priv-hint {
    background: var(--bg-elev);
    color: var(--text-dim);
    font-size: 0.9rem;
  }
  .loading { padding: 2rem; text-align: center; color: var(--text-dim); }
  footer {
    max-width: 880px;
    margin: 0.5rem auto 1.5rem;
    padding: 0 1.25rem;
    display: flex;
    justify-content: space-between;
    color: var(--text-dim);
    font-size: 0.82rem;
  }
  footer a { color: var(--text-dim); }
</style>
