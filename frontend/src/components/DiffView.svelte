<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '../lib/api';
  import type { DiffResponse, HistorySummary } from '../lib/types';

  let { initialA = null }: { initialA?: string | null } = $props();

  let history = $state<HistorySummary[]>([]);
  // Seeded once at mount from the prop. Parent uses {#key} to force a remount
  // when it wants to seed a different scan, so reading the prop once here is
  // intentional (avoiding the Svelte state_referenced_locally warning).
  let a = $state<string>('');
  let b = $state<string>('');
  let loading = $state(false);
  let error = $state<string | null>(null);
  let result = $state<DiffResponse | null>(null);

  onMount(async () => {
    a = initialA ?? '';
    try {
      history = await api.history();
      // Auto-pick a sensible second scan: the most recent one with matching
      // targets that isn't already `a`.
      if (a && !b) {
        const me = history.find((h) => h.id === a);
        if (me) {
          const candidate = history.find(
            (h) => h.id !== a && shareTarget(h.targets, me.targets),
          );
          if (candidate) b = candidate.id;
        }
      }
    } catch (e) {
      error = (e as Error).message;
    }
  });

  function shareTarget(x: string[], y: string[]): boolean {
    return x.some((t) => y.includes(t));
  }

  async function run() {
    if (!a || !b) return;
    if (a === b) {
      error = 'Pick two different scans.';
      return;
    }
    loading = true;
    error = null;
    try {
      result = await api.diff(a, b);
    } catch (e) {
      error = (e as Error).message;
    } finally {
      loading = false;
    }
  }

  function fmt(t: string): string {
    return new Date(t).toLocaleString();
  }
</script>

<section>
  <div class="head">
    <h2>Compare two scans</h2>
  </div>
  <p class="hint">
    Pick two scans of the same targets to see what changed between them — new hosts, new open ports, service-version changes, status flips.
  </p>

  <div class="pickers">
    <label>
      <span class="lbl">A (older)</span>
      <select bind:value={a}>
        <option value="">— pick a scan —</option>
        {#each history as h (h.id)}
          <option value={h.id}>{fmt(h.started)} · {h.targets.join(', ') || '(no targets)'}</option>
        {/each}
      </select>
    </label>
    <label>
      <span class="lbl">B (newer)</span>
      <select bind:value={b}>
        <option value="">— pick a scan —</option>
        {#each history as h (h.id)}
          <option value={h.id}>{fmt(h.started)} · {h.targets.join(', ') || '(no targets)'}</option>
        {/each}
      </select>
    </label>
    <button class="primary" onclick={run} disabled={loading || !a || !b}>
      {loading ? 'Diffing…' : 'Compare'}
    </button>
  </div>

  {#if error}<div class="error">{error}</div>{/if}

  {#if result}
    {@const d = result.diff}
    {#if d.identical}
      <div class="ok-box">No differences. Both scans report identical hosts, ports, and service versions.</div>
    {:else}
      <div class="meta">
        <span class="pill">A · {fmt(result.a_started)}</span>
        <span class="pill">B · {fmt(result.b_started)}</span>
      </div>

      {#if d.hosts_added && d.hosts_added.length > 0}
        <h3>+ hosts that appeared</h3>
        <ul class="changes added">
          {#each d.hosts_added as h (h.address)}
            <li><code>{h.address}</code>{#if h.hostname} <span class="hn">{h.hostname}</span>{/if}</li>
          {/each}
        </ul>
      {/if}

      {#if d.hosts_removed && d.hosts_removed.length > 0}
        <h3>− hosts that disappeared</h3>
        <ul class="changes removed">
          {#each d.hosts_removed as h (h.address)}
            <li><code>{h.address}</code>{#if h.hostname} <span class="hn">{h.hostname}</span>{/if}</li>
          {/each}
        </ul>
      {/if}

      {#if d.hosts_changed && d.hosts_changed.length > 0}
        <h3>Δ hosts that changed</h3>
        {#each d.hosts_changed as hc (hc.address)}
          <article class="host-change">
            <header>
              <code>{hc.address}</code>
              {#if hc.hostname}<span class="hn">{hc.hostname}</span>{/if}
              {#if hc.status_before || hc.status_after}
                <span class="pill warn">status: {hc.status_before ?? '—'} → {hc.status_after ?? '—'}</span>
              {/if}
              {#if hc.os_before || hc.os_after}
                <span class="pill">OS: {hc.os_before || '—'} → {hc.os_after || '—'}</span>
              {/if}
            </header>
            {#if hc.ports_added && hc.ports_added.length > 0}
              <div class="line added">
                <span class="tag">+ open</span>
                {#each hc.ports_added as p, i (p.protocol + p.portid)}
                  {#if i > 0}, {/if}
                  <code>{p.protocol}/{p.portid}</code>{#if p.service} <span class="svc">{p.service}{p.version ? ` ${p.version}` : ''}</span>{/if}
                {/each}
              </div>
            {/if}
            {#if hc.ports_removed && hc.ports_removed.length > 0}
              <div class="line removed">
                <span class="tag">− gone</span>
                {#each hc.ports_removed as p, i (p.protocol + p.portid)}
                  {#if i > 0}, {/if}
                  <code>{p.protocol}/{p.portid}</code>{#if p.service} <span class="svc">{p.service}{p.version ? ` ${p.version}` : ''}</span>{/if}
                {/each}
              </div>
            {/if}
            {#if hc.ports_changed && hc.ports_changed.length > 0}
              <ul class="ports-changed">
                {#each hc.ports_changed as pc (pc.protocol + pc.portid)}
                  <li>
                    <code>{pc.protocol}/{pc.portid}</code>
                    {#if pc.state_before || pc.state_after}
                      <span class="kv">state: {pc.state_before ?? '—'} → {pc.state_after ?? '—'}</span>
                    {/if}
                    {#if pc.service_before !== pc.service_after}
                      <span class="kv">service: {pc.service_before || '—'} → {pc.service_after || '—'}</span>
                    {/if}
                    {#if pc.version_before !== pc.version_after}
                      <span class="kv">version: {pc.version_before || '—'} → {pc.version_after || '—'}</span>
                    {/if}
                  </li>
                {/each}
              </ul>
            {/if}
          </article>
        {/each}
      {/if}
    {/if}
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
  .head { margin-bottom: 0.4rem; }
  h2 { margin: 0; font-size: 1.05rem; }
  .hint { margin: 0.2rem 0 0.85rem; color: var(--text-dim); font-size: 0.88rem; }

  .pickers {
    display: flex;
    gap: 0.55rem;
    align-items: end;
    flex-wrap: wrap;
    margin-bottom: 0.75rem;
  }
  .pickers label { display: flex; flex-direction: column; gap: 0.2rem; min-width: 16rem; flex: 1 1 16rem; }
  .lbl { color: var(--text-dim); font-size: 0.8rem; }
  select {
    background: var(--bg-elev);
    color: var(--text);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
    padding: 0.4rem 0.55rem;
    font: inherit;
  }

  .error {
    padding: 0.5rem 0.75rem;
    background: rgba(220, 38, 38, 0.08);
    border: 1px solid rgba(220, 38, 38, 0.2);
    color: var(--danger);
    border-radius: var(--radius);
    margin: 0.4rem 0;
  }
  .ok-box {
    padding: 0.65rem 0.85rem;
    background: rgba(21, 128, 61, 0.08);
    border: 1px solid rgba(21, 128, 61, 0.25);
    border-radius: var(--radius);
    color: var(--ok);
    margin-top: 0.6rem;
  }

  .meta { display: flex; gap: 0.4rem; margin: 0.7rem 0; }

  h3 { margin: 0.85rem 0 0.4rem; font-size: 0.92rem; color: var(--text-dim); font-weight: 600; }
  ul.changes { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 0.2rem; }
  ul.changes li {
    padding: 0.3rem 0.6rem;
    border-radius: var(--radius);
    border: 1px solid var(--border);
  }
  ul.changes.added li { border-color: rgba(21,128,61,0.25); background: rgba(21,128,61,0.06); }
  ul.changes.removed li { border-color: rgba(220,38,38,0.25); background: rgba(220,38,38,0.06); }
  .hn { color: var(--text-dim); margin-left: 0.4rem; }

  .host-change {
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 0.55rem 0.75rem;
    margin: 0.5rem 0;
    background: var(--bg);
  }
  .host-change header { display: flex; flex-wrap: wrap; gap: 0.4rem; align-items: center; margin-bottom: 0.35rem; }
  .line { font-size: 0.86rem; padding: 0.25rem 0; }
  .line .tag { font-weight: 600; margin-right: 0.4rem; }
  .line.added { color: var(--ok); }
  .line.removed { color: var(--danger); }
  .svc { color: var(--text-dim); }

  ul.ports-changed { list-style: none; margin: 0.3rem 0 0; padding: 0; }
  ul.ports-changed li {
    padding: 0.2rem 0;
    border-top: 1px dashed var(--border);
    font-size: 0.85rem;
    display: flex; gap: 0.6rem; flex-wrap: wrap;
  }
  ul.ports-changed li:first-child { border-top: none; }
  .kv { color: var(--text-dim); }
</style>
