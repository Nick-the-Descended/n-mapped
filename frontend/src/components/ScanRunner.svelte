<script lang="ts">
  import { onMount } from 'svelte';
  import { api, streamScan } from '../lib/api';
  import { EVT, on } from '../lib/events';
  import type { Host, RunStats, ScanInfo, ScanRequest, TaskProgress } from '../lib/types';

  let {
    request,
    onStart,
    onUpdate,
    onDone,
  }: {
    request: ScanRequest;
    onStart: (id: string, display: string) => void;
    onUpdate: (state: {
      hosts: Host[];
      runstats: RunStats | null;
      scaninfos: ScanInfo[];
      progress: TaskProgress | null;
      stderr: string[];
    }) => void;
    onDone: (id: string, exitCode: number, error: string | null) => void;
  } = $props();

  let starting = $state(false);
  let scanID = $state<string | null>(null);
  let display = $state<string | null>(null);
  let finished = $state(false);
  let exitCode = $state<number | null>(null);
  let errorMsg = $state<string | null>(null);
  let stop: (() => void) | null = null;

  // Local accumulators pushed up via onUpdate. Marked $state so Svelte 5
  // doesn't warn on mutation; only stderr is read in this component's template.
  let hosts = $state<Host[]>([]);
  let scaninfos = $state<ScanInfo[]>([]);
  let runstats = $state<RunStats | null>(null);
  let progress = $state<TaskProgress | null>(null);
  let stderr = $state<string[]>([]);

  function pushUpdate() {
    onUpdate({ hosts: [...hosts], scaninfos: [...scaninfos], runstats, progress, stderr: [...stderr] });
  }

  async function start() {
    errorMsg = null;
    finished = false;
    exitCode = null;
    hosts = []; scaninfos = []; runstats = null; progress = null; stderr = [];
    pushUpdate();
    starting = true;
    try {
      const res = await api.startScan(request);
      scanID = res.id;
      display = res.display;
      onStart(res.id, res.display);
      stop = streamScan(res.id, (ev) => {
        switch (ev.kind) {
          case 'host':         if (ev.host) { hosts.push(ev.host); pushUpdate(); } break;
          case 'scaninfo':     if (ev.scaninfo) { scaninfos.push(ev.scaninfo); pushUpdate(); } break;
          case 'taskprogress': if (ev.progress) { progress = ev.progress; pushUpdate(); } break;
          case 'runstats':     if (ev.runstats) { runstats = ev.runstats; pushUpdate(); } break;
          case 'stderr':       if (ev.line) { stderr.push(ev.line); pushUpdate(); } break;
          case 'done':
            finished = true;
            exitCode = ev.code ?? 0;
            onDone(scanID!, exitCode, ev.err ?? null);
            break;
          case 'error':
            finished = true;
            errorMsg = ev.err ?? 'unknown error';
            onDone(scanID!, exitCode ?? -1, errorMsg);
            break;
        }
      });
    } catch (e) {
      errorMsg = (e as Error).message;
    } finally {
      starting = false;
    }
  }

  async function cancel() {
    if (!scanID) return;
    try { await api.stopScan(scanID); }
    catch (e) { errorMsg = (e as Error).message; }
  }

  function clear() {
    if (stop) stop();
    stop = null;
    scanID = null;
    display = null;
    finished = false;
    exitCode = null;
    errorMsg = null;
    hosts = []; scaninfos = []; runstats = null; progress = null; stderr = [];
    pushUpdate();
  }

  // Keyboard shortcut hooks: ⌘⏎ runs (when no scan in flight); ⌘. stops.
  onMount(() => {
    const offRun = on(EVT.runScan, () => { if (!scanID && !starting) start(); });
    const offStop = on(EVT.stopScan, () => { if (scanID && !finished) cancel(); });
    return () => { offRun(); offStop(); };
  });
</script>

<section>
  <div class="actions">
    {#if !scanID}
      <button class="primary" onclick={start} disabled={starting}>
        {starting ? 'Starting…' : 'Run scan'}
      </button>
    {:else if !finished}
      <button class="danger" onclick={cancel}>Stop scan</button>
    {:else}
      <button onclick={clear}>Clear &amp; new scan</button>
    {/if}
    {#if scanID}
      <span class="pill {finished ? (exitCode === 0 ? 'ok' : 'danger') : 'accent'}">
        {finished
          ? (exitCode === 0 ? `done (exit ${exitCode})` : `failed (exit ${exitCode})`)
          : 'running…'}
      </span>
      <span class="scan-id">id: <code>{scanID.slice(0, 8)}</code></span>
    {/if}
  </div>

  {#if errorMsg}
    <div class="error">⚠️ {errorMsg}</div>
  {/if}

  {#if scanID && stderr.length > 0}
    <details class="stderr">
      <summary>nmap status / warnings ({stderr.length})</summary>
      <pre>{stderr.join('\n')}</pre>
    </details>
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
  details.stderr { margin-top: 0.85rem; }
  details.stderr summary { cursor: pointer; color: var(--text-dim); font-size: 0.85rem; }
  details.stderr pre {
    max-height: 14rem;
    overflow: auto;
    margin-top: 0.4rem;
    font-size: 0.8rem;
  }
</style>
