<script lang="ts">
  type Tab = 'builder' | 'results' | 'history' | 'diff';
  let {
    active = $bindable<Tab>('builder'),
    counts,
  }: {
    active: Tab;
    counts: { results: number; history: number };
  } = $props();

  const tabs: Array<{ id: Tab; label: string }> = [
    { id: 'builder', label: 'Builder' },
    { id: 'results', label: 'Results' },
    { id: 'history', label: 'History' },
    { id: 'diff',    label: 'Diff' },
  ];
</script>

<nav>
  {#each tabs as t (t.id)}
    <button
      class:active={active === t.id}
      onclick={() => (active = t.id)}
    >
      {t.label}
      {#if t.id === 'results' && counts.results > 0}
        <span class="badge">{counts.results}</span>
      {:else if t.id === 'history' && counts.history > 0}
        <span class="badge">{counts.history}</span>
      {/if}
    </button>
  {/each}
</nav>

<style>
  nav {
    display: flex;
    gap: 0.25rem;
    padding: 0 1.25rem;
    border-bottom: 1px solid var(--border);
    background: var(--bg-elev);
  }
  button {
    background: transparent;
    border: none;
    border-radius: 0;
    border-bottom: 2px solid transparent;
    padding: 0.7rem 1rem;
    color: var(--text-dim);
    margin-bottom: -1px;
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
  }
  button:hover { color: var(--text); border-color: var(--border-strong); }
  button.active {
    color: var(--accent);
    border-color: var(--accent);
  }
  .badge {
    background: var(--code-bg);
    border-radius: 999px;
    padding: 0.05em 0.55em;
    font-size: 0.75rem;
  }
  button.active .badge {
    background: var(--accent);
    color: var(--accent-fg);
  }
</style>
