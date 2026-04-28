<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '../lib/api';
  import type { HistoryRecord, HistorySummary } from '../lib/types';

  let {
    onOpen,
    onRerun,
  }: {
    onOpen: (rec: HistoryRecord) => void;
    onRerun: (rec: HistoryRecord) => void;
  } = $props();

  let entries = $state<HistorySummary[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);

  async function refresh() {
    loading = true;
    error = null;
    try {
      entries = await api.history();
    } catch (e) {
      error = (e as Error).message;
    } finally {
      loading = false;
    }
  }

  onMount(refresh);

  async function open(s: HistorySummary) {
    try {
      const rec = await api.historyRecord(s.id);
      onOpen(rec);
    } catch (e) {
      error = (e as Error).message;
    }
  }

  async function rerun(s: HistorySummary, ev: Event) {
    ev.stopPropagation();
    try {
      const rec = await api.historyRecord(s.id);
      onRerun(rec);
    } catch (e) {
      error = (e as Error).message;
    }
  }

  async function remove(s: HistorySummary, ev: Event) {
    ev.stopPropagation();
    if (!confirm(`Delete scan ${s.id.slice(0, 8)}? This is permanent.`)) return;
    try {
      await api.historyDelete(s.id);
      entries = entries.filter((e) => e.id !== s.id);
    } catch (e) {
      error = (e as Error).message;
    }
  }

  function fmtTime(iso: string): string {
    const d = new Date(iso);
    return d.toLocaleString();
  }

  function fmtDuration(start: string, end: string): string {
    const ms = new Date(end).getTime() - new Date(start).getTime();
    if (ms < 1000) return `${ms} ms`;
    if (ms < 60_000) return `${(ms / 1000).toFixed(1)} s`;
    return `${(ms / 60_000).toFixed(1)} min`;
  }
</script>

<section>
  <div class="head">
    <h2>Scan history</h2>
    <button onclick={refresh} disabled={loading}>{loading ? 'Refreshing…' : 'Refresh'}</button>
  </div>

  {#if error}
    <div class="error">{error}</div>
  {/if}

  {#if !loading && entries.length === 0}
    <p class="empty">No scans yet. Run one from the Builder tab — it'll show up here automatically.</p>
  {/if}

  {#each entries as s (s.id)}
    <div
      class="row"
      role="button"
      tabindex="0"
      onclick={() => open(s)}
      onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); open(s); } }}
    >
      <div class="row-main">
        <div class="row-top">
          <span class="targets">{s.targets.join(', ') || '(no targets)'}</span>
          <span class="when">{fmtTime(s.started)}</span>
        </div>
        <div class="row-cmd">{s.display}</div>
        <div class="row-stats">
          <span class="pill {s.exit_code === 0 ? 'ok' : 'danger'}">exit {s.exit_code}</span>
          <span class="pill">{fmtDuration(s.started, s.ended)}</span>
          {#if s.hosts_total > 0}
            <span class="pill ok">{s.hosts_up}/{s.hosts_total} up</span>
          {/if}
          {#if s.open_ports > 0}
            <span class="pill accent">{s.open_ports} open ports</span>
          {/if}
          {#if s.error}
            <span class="pill danger" title={s.error}>error</span>
          {/if}
        </div>
      </div>
      <div class="row-actions">
        <button onclick={(e) => rerun(s, e)} title="Re-run with the same options">Re-run</button>
        <a class="dl" href={api.historyXmlURL(s.id)} download onclick={(e) => e.stopPropagation()}>XML</a>
        <button onclick={(e) => remove(s, e)} title="Delete this record">Del</button>
      </div>
    </div>
  {/each}
</section>

<style>
  section {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 1rem 1.1rem;
    box-shadow: var(--shadow);
  }
  .head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.65rem;
  }
  h2 { margin: 0; font-size: 1.05rem; }
  .empty { color: var(--text-dim); }
  .error {
    padding: 0.5rem 0.75rem;
    margin-bottom: 0.6rem;
    background: rgba(220, 38, 38, 0.08);
    border: 1px solid rgba(220, 38, 38, 0.2);
    color: var(--danger);
    border-radius: var(--radius);
  }
  .row {
    display: flex;
    justify-content: space-between;
    align-items: stretch;
    width: 100%;
    text-align: left;
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 0.6rem 0.85rem;
    margin-top: 0.5rem;
    gap: 0.75rem;
  }
  .row:hover { border-color: var(--accent); }
  .row-main { flex: 1; min-width: 0; }
  .row-top {
    display: flex;
    justify-content: space-between;
    gap: 0.5rem;
    align-items: baseline;
  }
  .targets { font-weight: 600; word-break: break-word; }
  .when { color: var(--text-dim); font-size: 0.8rem; }
  .row-cmd {
    font-family: ui-monospace, monospace;
    font-size: 0.8rem;
    color: var(--text-dim);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    margin: 0.2rem 0;
  }
  .row-stats { display: flex; gap: 0.3rem; flex-wrap: wrap; }
  .row-actions {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    align-self: center;
  }
  .row-actions button, .row-actions .dl {
    padding: 0.25rem 0.55rem;
    font-size: 0.78rem;
  }
  .dl {
    display: inline-block;
    text-align: center;
    background: var(--bg-elev);
    color: var(--text);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
    text-decoration: none;
  }
  .dl:hover { border-color: var(--accent); }
</style>
