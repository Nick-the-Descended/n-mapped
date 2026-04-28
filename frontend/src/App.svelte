<script lang="ts">
  import { onMount } from 'svelte';
  import Header from './components/Header.svelte';
  import TabBar from './components/TabBar.svelte';
  import TargetInput from './components/TargetInput.svelte';
  import FlagPicker from './components/FlagPicker.svelte';
  import CommandPreview from './components/CommandPreview.svelte';
  import ScanRunner from './components/ScanRunner.svelte';
  import ResultsView from './components/ResultsView.svelte';
  import HistoryPanel from './components/HistoryPanel.svelte';
  import Glossary from './components/Glossary.svelte';
  import EthicalUseSplash from './components/EthicalUseSplash.svelte';
  import ReverseParser from './components/ReverseParser.svelte';
  import { api } from './lib/api';
  import type {
    Catalog, Host, HistoryRecord, NmapInfo, PrivilegeState, RunStats,
    ScanInfo, ScanRequest, TaskProgress,
  } from './lib/types';
  import { previewCommand, previewSummary } from './lib/store';

  // Boot state.
  let catalog = $state<Catalog | null>(null);
  let nmap = $state<NmapInfo | null>(null);
  let privilege = $state<PrivilegeState | null>(null);
  let bootError = $state<string | null>(null);

  // Builder state.
  let targets = $state('');
  let selected = $state<Set<string>>(new Set());
  let values = $state<Record<string, string>>({});

  // Tab navigation.
  let activeTab = $state<'builder' | 'results' | 'history'>('builder');

  // Currently displayed result — either the live scan or a loaded historical one.
  type ResultSource = { kind: 'live'; id: string | null; display: string | null }
                    | { kind: 'history'; record: HistoryRecord };
  let resultSource = $state<ResultSource>({ kind: 'live', id: null, display: null });

  // Live scan state (updated by ScanRunner).
  let liveHosts = $state<Host[]>([]);
  let liveRunstats = $state<RunStats | null>(null);
  let liveScanInfos = $state<ScanInfo[]>([]);
  let liveProgress = $state<TaskProgress | null>(null);

  // History refresh trigger — bumped when a scan completes so HistoryPanel can pick it up.
  let historyRefreshKey = $state(0);

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

  let resultsCount = $derived.by(() => {
    if (resultSource.kind === 'history') {
      return resultSource.record.result?.run?.hosts?.length ?? 0;
    }
    return liveHosts.length;
  });

  function handleScanStart(_id: string, display: string) {
    resultSource = { kind: 'live', id: _id, display };
    activeTab = 'results';
  }

  function handleScanUpdate(state: { hosts: Host[]; runstats: RunStats | null; scaninfos: ScanInfo[]; progress: TaskProgress | null; }) {
    if (resultSource.kind !== 'live') return;
    liveHosts = state.hosts;
    liveRunstats = state.runstats;
    liveScanInfos = state.scaninfos;
    liveProgress = state.progress;
  }

  function handleScanDone() {
    historyRefreshKey++;
  }

  function applyReverseParse(r: { targets: string; flagIDs: string[]; flagValues: Record<string, string> }) {
    targets = r.targets;
    selected = new Set(r.flagIDs);
    values = r.flagValues;
  }

  function openHistoryRecord(rec: HistoryRecord) {
    resultSource = { kind: 'history', record: rec };
    activeTab = 'results';
  }

  function rerunFromRecord(rec: HistoryRecord) {
    targets = (rec.targets ?? []).join(' ');
    selected = new Set(rec.flag_ids ?? []);
    // Inline values aren't stored separately yet; user re-enters if needed.
    values = {};
    activeTab = 'builder';
  }

  // Surface the raw XML download URL when looking at a historical record.
  let historyXmlURL = $derived(
    resultSource.kind === 'history' ? api.historyXmlURL(resultSource.record.id) : null,
  );
</script>

<EthicalUseSplash />

<Header {nmap} {privilege} />

<TabBar bind:active={activeTab} counts={{ results: resultsCount, history: historyRefreshKey }} />

<main>
  {#if bootError}
    <div class="boot-error">
      Failed to load backend: <code>{bootError}</code>.
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

    {#if activeTab === 'builder'}
      <ReverseParser flags={catalog.flags} onApply={applyReverseParse} />
      <TargetInput bind:value={targets} />
      <FlagPicker {catalog} {privilege} bind:selected bind:values />
      <CommandPreview {command} {summary} />
      <ScanRunner
        {request}
        onStart={handleScanStart}
        onUpdate={handleScanUpdate}
        onDone={handleScanDone}
      />
      <Glossary />
    {:else if activeTab === 'results'}
      {#if resultSource.kind === 'live' && resultSource.id}
        <ResultsView
          hosts={liveHosts}
          runstats={liveRunstats}
          scaninfos={liveScanInfos}
          progress={liveProgress}
          label={`Live scan ${resultSource.id.slice(0, 8)}`}
        />
      {:else if resultSource.kind === 'history'}
        <ResultsView
          hosts={resultSource.record.result?.run?.hosts ?? []}
          runstats={resultSource.record.result?.run?.runstats ?? null}
          scaninfos={resultSource.record.result?.run?.scaninfo ?? []}
          progress={null}
          label={`From history · ${resultSource.record.id.slice(0, 8)}`}
        />
        {#if historyXmlURL}
          <p class="dl-row">
            <a href={historyXmlURL} download>Download raw nmap XML</a>
          </p>
        {/if}
      {:else}
        <div class="empty-state">
          No result loaded yet. Run a scan from the Builder tab, or pick one from History.
        </div>
      {/if}
    {:else if activeTab === 'history'}
      {#key historyRefreshKey}
        <HistoryPanel onOpen={openHistoryRecord} onRerun={rerunFromRecord} />
      {/key}
    {/if}
  {/if}
</main>

<footer>
  <span>nmap on {nmap?.path ?? 'PATH'}</span>
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
  .loading, .empty-state {
    padding: 2rem;
    text-align: center;
    color: var(--text-dim);
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
  }
  .dl-row { margin: 0.5rem 0 0; }
  .dl-row a {
    display: inline-block;
    padding: 0.4rem 0.75rem;
    background: var(--bg-elev);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
    color: var(--text);
    text-decoration: none;
  }
  .dl-row a:hover { border-color: var(--accent); color: var(--accent); }
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
