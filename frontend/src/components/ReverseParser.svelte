<script lang="ts">
  import type { Flag } from '../lib/types';
  import { parseCommand } from '../lib/store';

  let {
    flags,
    onApply,
  }: {
    flags: Flag[];
    onApply: (r: { targets: string; flagIDs: string[]; flagValues: Record<string, string> }) => void;
  } = $props();

  let input = $state('');
  let preview = $derived.by(() => {
    if (!input.trim()) return null;
    return parseCommand(input, flags);
  });

  function apply() {
    if (!preview) return;
    onApply({
      targets: preview.targets.join(' '),
      flagIDs: preview.flagIDs,
      flagValues: preview.flagValues,
    });
    input = '';
  }
</script>

<details class="rev">
  <summary>Paste an existing nmap command &mdash; we'll fill in the builder</summary>
  <div class="body">
    <input
      type="text"
      placeholder="e.g. nmap -sS -sV -p 22,80 192.168.1.0/24"
      bind:value={input}
    />
    {#if preview}
      <div class="preview">
        <div class="row">
          <span class="label">Targets</span>
          <span>{preview.targets.length ? preview.targets.join(', ') : '(none)'}</span>
        </div>
        <div class="row">
          <span class="label">Recognized</span>
          <span>{preview.flagIDs.length ? preview.flagIDs.join(', ') : '(none)'}</span>
        </div>
        {#if preview.unrecognized.length > 0}
          <div class="row warn">
            <span class="label">Unknown tokens</span>
            <span>{preview.unrecognized.join(' ')}</span>
          </div>
        {/if}
        <button class="primary" onclick={apply} disabled={preview.flagIDs.length === 0 && preview.targets.length === 0}>
          Apply to builder
        </button>
      </div>
    {/if}
  </div>
</details>

<style>
  .rev {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
  }
  summary {
    cursor: pointer;
    padding: 0.85rem 1.1rem;
    color: var(--text-dim);
    font-weight: 600;
    list-style: none;
  }
  summary::before { content: '▸ '; color: var(--text-dim); }
  details[open] > summary::before { content: '▾ '; }
  summary::-webkit-details-marker { display: none; }
  .body { padding: 0 1.1rem 1rem; display: flex; flex-direction: column; gap: 0.75rem; }
  .preview { display: flex; flex-direction: column; gap: 0.4rem; }
  .row { display: flex; gap: 0.6rem; }
  .row.warn { color: var(--warn); }
  .label {
    color: var(--text-dim);
    font-weight: 600;
    min-width: 7rem;
  }
</style>
