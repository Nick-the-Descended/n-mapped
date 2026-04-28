<script lang="ts">
  import ProgressBar from './ProgressBar.svelte';
  import type { Host, RunStats, ScanInfo, TaskProgress } from '../lib/types';

  let {
    hosts,
    runstats,
    scaninfos,
    progress,
    label,
  }: {
    hosts: Host[];
    runstats: RunStats | null;
    scaninfos: ScanInfo[];
    progress: TaskProgress | null;
    label?: string;
  } = $props();

  let filter = $state('');
  let openOnly = $state(true);

  function primaryAddr(h: Host): string {
    const ip = h.addresses.find((a) => a.addrtype === 'ipv4' || a.addrtype === 'ipv6');
    return (ip?.addr) ?? h.addresses[0]?.addr ?? '?';
  }

  function hostName(h: Host): string {
    return h.hostnames?.hostname?.[0]?.name ?? '';
  }

  function visibleHosts(): Host[] {
    const q = filter.trim().toLowerCase();
    return hosts.filter((h) => {
      if (q) {
        const hay = [primaryAddr(h), hostName(h),
          ...(h.ports?.ports?.flatMap((p) => [String(p.portid), p.service?.name ?? '', p.service?.product ?? '']) ?? []),
        ].join(' ').toLowerCase();
        if (!hay.includes(q)) return false;
      }
      return true;
    });
  }

  function visiblePorts(h: Host) {
    const ports = h.ports?.ports ?? [];
    return openOnly ? ports.filter((p) => p.state.state === 'open') : ports;
  }
</script>

{#if label}
  <div class="title">{label}</div>
{/if}

<section>
  <div class="meta">
    {#if runstats}
      <span class="pill ok">{runstats.hosts.up} up</span>
      <span class="pill">{runstats.hosts.total} scanned</span>
      <span class="pill">elapsed {runstats.finished.elapsed.toFixed(1)}s</span>
    {:else if hosts.length > 0}
      <span class="pill accent">{hosts.length} host{hosts.length === 1 ? '' : 's'} so far…</span>
    {/if}
    {#each scaninfos as si (si.type + si.protocol)}
      <span class="pill">{si.protocol.toUpperCase()} {si.type}</span>
    {/each}
  </div>

  <ProgressBar progress={progress} finished={runstats !== null} />

  <div class="filters">
    <input type="text" placeholder="Filter hosts / ports / services…" bind:value={filter} />
    <label class="check">
      <input type="checkbox" bind:checked={openOnly} /> open ports only
    </label>
  </div>

  {#if hosts.length === 0}
    <p class="empty">No hosts yet. Output appears here as nmap reports each one.</p>
  {/if}

  {#each visibleHosts() as h (primaryAddr(h) + (h.starttime ?? 0))}
    {@const ports = visiblePorts(h)}
    <article class="host">
      <header>
        <div class="addr">
          <span class="ip">{primaryAddr(h)}</span>
          {#if hostName(h)}<span class="hostname">{hostName(h)}</span>{/if}
        </div>
        <div class="badges">
          <span class="pill {h.status.state === 'up' ? 'ok' : ''}">{h.status.state}</span>
          {#if h.status.reason}
            <span class="pill" title="reason">{h.status.reason}</span>
          {/if}
          {#if h.os?.matches?.[0]}
            <span class="pill accent" title="OS guess">
              {h.os.matches[0].name} ({h.os.matches[0].accuracy}%)
            </span>
          {/if}
        </div>
      </header>

      {#if ports.length > 0}
        <table>
          <thead>
            <tr>
              <th>Port</th>
              <th>Proto</th>
              <th>State</th>
              <th>Service</th>
              <th>Version</th>
            </tr>
          </thead>
          <tbody>
            {#each ports as p (p.protocol + p.portid)}
              <tr>
                <td class="num">{p.portid}</td>
                <td>{p.protocol}</td>
                <td>
                  <span class="pill {p.state.state === 'open' ? 'ok' : p.state.state === 'closed' ? '' : 'warn'}">
                    {p.state.state}
                  </span>
                </td>
                <td>{p.service?.name ?? '—'}</td>
                <td>
                  {#if p.service?.product}
                    {p.service.product}{p.service.version ? ` ${p.service.version}` : ''}
                  {:else if p.service?.extrainfo}
                    {p.service.extrainfo}
                  {:else}
                    —
                  {/if}
                  {#if p.scripts && p.scripts.length > 0}
                    <details class="scripts">
                      <summary>{p.scripts.length} script result{p.scripts.length === 1 ? '' : 's'}</summary>
                      {#each p.scripts as sc (sc.id)}
                        <div class="script">
                          <code>{sc.id}</code>
                          <pre>{sc.output}</pre>
                        </div>
                      {/each}
                    </details>
                  {/if}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      {:else}
        <p class="muted">{openOnly ? 'No open ports (toggle "open ports only" to see closed/filtered).' : 'No ports reported.'}</p>
      {/if}

      {#if h.ports?.extraports && h.ports.extraports.length > 0}
        <p class="extras">
          {#each h.ports.extraports as ep (ep.state)}
            <span class="pill">{ep.count} {ep.state}</span>
          {/each}
        </p>
      {/if}
    </article>
  {/each}
</section>

<style>
  .title {
    font-weight: 600;
    margin-bottom: 0.4rem;
    color: var(--text-dim);
  }
  section {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 1rem 1.1rem;
    box-shadow: var(--shadow);
  }
  .meta { display: flex; flex-wrap: wrap; gap: 0.4rem; margin-bottom: 0.85rem; }
  .filters {
    display: flex;
    gap: 0.6rem;
    align-items: center;
    margin-bottom: 0.85rem;
    flex-wrap: wrap;
  }
  .filters input { flex: 1 1 220px; }
  .check { display: inline-flex; align-items: center; gap: 0.35rem; color: var(--text-dim); font-size: 0.88rem; }
  .empty { color: var(--text-dim); margin: 0.5rem 0 0; }

  .host {
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 0.85rem 1rem;
    margin-top: 0.7rem;
    background: var(--bg);
  }
  .host header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-bottom: 0.6rem;
  }
  .addr { display: flex; align-items: baseline; gap: 0.5rem; }
  .ip { font-family: ui-monospace, monospace; font-weight: 600; }
  .hostname { color: var(--text-dim); font-size: 0.88rem; }
  .badges { display: flex; gap: 0.3rem; flex-wrap: wrap; }

  table { width: 100%; border-collapse: collapse; font-size: 0.88rem; }
  th { text-align: left; font-weight: 600; padding: 0.35rem 0.6rem; color: var(--text-dim); border-bottom: 1px solid var(--border); }
  td { padding: 0.35rem 0.6rem; vertical-align: top; border-bottom: 1px dashed var(--border); }
  td.num { font-family: ui-monospace, monospace; }
  tr:last-child td { border-bottom: none; }

  .scripts { margin-top: 0.3rem; }
  .scripts summary { cursor: pointer; color: var(--text-dim); font-size: 0.82rem; }
  .script { margin: 0.4rem 0; }
  .script code { font-size: 0.8rem; }
  .script pre {
    margin: 0.2rem 0 0; font-size: 0.82rem; max-height: 14rem; overflow: auto;
    white-space: pre-wrap;
  }

  .extras { display: flex; gap: 0.35rem; margin: 0.5rem 0 0; flex-wrap: wrap; }
  .muted { color: var(--text-dim); font-size: 0.88rem; margin: 0.3rem 0 0; }
</style>
