<script lang="ts">
  import { api, streamScan } from '../lib/api';
  import type { ScanEvent, ScanRequest, ScanStartResponse } from '../lib/types';

  let { request }: { request: ScanRequest } = $props();

  let starting = $state(false);
  let scan = $state<ScanStartResponse | null>(null);
  let events = $state<ScanEvent[]>([]);
  let finished = $state(false);
  let exitCode = $state<number | null>(null);
  let errorMsg = $state<string | null>(null);
  let stop: (() => void) | null = null;

  async function start() {
    errorMsg = null;
    events = [];
    finished = false;
    exitCode = null;
    starting = true;
    try {
      scan = await api.startScan(request);
      stop = streamScan(scan.id, (ev) => {
        events = [...events, ev];
        if (ev.kind === 'done') {
          finished = true;
          exitCode = ev.code ?? 0;
        } else if (ev.kind === 'error') {
          finished = true;
          errorMsg = ev.err ?? 'unknown error';
        }
      });
    } catch (e) {
      errorMsg = (e as Error).message;
    } finally {
      starting = false;
    }
  }

  async function cancel() {
    if (!scan) return;
    try {
      await api.stopScan(scan.id);
    } catch (e) {
      errorMsg = (e as Error).message;
    }
  }

  function clear() {
    if (stop) stop();
    stop = null;
    scan = null;
    events = [];
    finished = false;
    exitCode = null;
    errorMsg = null;
  }

  function formatLine(ev: ScanEvent): string {
    switch (ev.kind) {
      case 'start':  return '▶ scan started';
      case 'stdout': return ev.line ?? '';
      case 'stderr': return ev.line ?? '';
      case 'done':   return `■ done (exit ${ev.code ?? 0})`;
      case 'error':  return `✖ ${ev.err ?? 'error'}`;
      default:       return '';
    }
  }
</script>

<section>
  <div class="actions">
    {#if !scan}
      <button class="primary" onclick={start} disabled={starting}>
        {starting ? 'Starting…' : 'Run scan'}
      </button>
    {:else if !finished}
      <button class="danger" onclick={cancel}>Stop scan</button>
    {:else}
      <button onclick={clear}>Clear &amp; new scan</button>
    {/if}
    {#if scan}
      <span class="pill {finished ? (exitCode === 0 ? 'ok' : 'danger') : 'accent'}">
        {finished
          ? (exitCode === 0 ? `done (exit ${exitCode})` : `failed (exit ${exitCode})`)
          : 'running…'}
      </span>
      <span class="scan-id">id: <code>{scan.id.slice(0, 8)}</code></span>
    {/if}
  </div>

  {#if errorMsg}
    <div class="error">⚠️ {errorMsg}</div>
  {/if}

  {#if scan}
    <div class="output-label">Live output</div>
    <pre class="output">{#each events as ev (ev.when + ev.kind + (ev.line ?? ''))}<span class={ev.kind}>{formatLine(ev)}</span>
{/each}</pre>
  {/if}
</section>

<style>
  section {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 1rem 1.1rem;
    box-shadow: var(--shadow);
  }
  .actions {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    flex-wrap: wrap;
  }
  .scan-id { color: var(--text-dim); font-size: 0.82rem; }
  .error {
    margin-top: 0.6rem;
    padding: 0.5rem 0.75rem;
    background: rgba(220, 38, 38, 0.08);
    border: 1px solid rgba(220, 38, 38, 0.2);
    border-radius: var(--radius);
    color: var(--danger);
  }
  .output-label {
    margin: 0.85rem 0 0.35rem;
    font-weight: 600;
    color: var(--text-dim);
    font-size: 0.85rem;
  }
  .output {
    max-height: 22rem;
    overflow: auto;
    background: var(--code-bg);
    font-size: 0.82rem;
    line-height: 1.45;
  }
  .output .stderr { color: var(--warn); }
  .output .start, .output .done { color: var(--accent); }
  .output .error { color: var(--danger); }
</style>
