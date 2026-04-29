<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '../lib/api';
  import { EVT, on } from '../lib/events';
  import type { Favorite, Profile } from '../lib/types';

  let {
    profiles,
    currentTargets,
    currentFlagIDs,
    currentFlagValues,
    currentScriptIDs,
    currentScriptArgs,
    onApply,
  }: {
    profiles: Profile[];
    currentTargets: string;
    currentFlagIDs: string[];
    currentFlagValues: Record<string, string>;
    currentScriptIDs: string[];
    currentScriptArgs: Record<string, string>;
    onApply: (state: {
      targets: string;
      flagIDs: string[];
      flagValues: Record<string, string>;
      scriptIDs: string[];
      scriptArgs: Record<string, string>;
    }) => void;
  } = $props();

  let favorites = $state<Favorite[]>([]);
  let saving = $state(false);
  let savePrompt = $state(false);
  let saveName = $state('');
  let saveDesc = $state('');
  let error = $state<string | null>(null);

  async function loadFavorites() {
    try { favorites = await api.favorites(); }
    catch (e) { error = (e as Error).message; }
  }

  onMount(() => {
    loadFavorites();
    return on(EVT.saveFavorite, () => { if (canSave) savePrompt = true; });
  });

  function applyProfile(p: Profile) {
    onApply({
      targets: currentTargets,
      flagIDs: p.flag_ids ?? [],
      flagValues: { ...(p.flag_values ?? {}) },
      scriptIDs: p.script_ids ?? [],
      scriptArgs: { ...(p.script_args ?? {}) },
    });
  }

  function applyFavorite(f: Favorite) {
    onApply({
      targets: f.targets ?? '',
      flagIDs: f.flag_ids ?? [],
      flagValues: { ...(f.flag_values ?? {}) },
      scriptIDs: f.script_ids ?? [],
      scriptArgs: { ...(f.script_args ?? {}) },
    });
  }

  async function saveCurrent() {
    if (!saveName.trim()) return;
    saving = true;
    error = null;
    try {
      await api.saveFavorite({
        name: saveName.trim(),
        description: saveDesc.trim() || undefined,
        targets: currentTargets,
        flag_ids: currentFlagIDs,
        flag_values: currentFlagValues,
        script_ids: currentScriptIDs,
        script_args: currentScriptArgs,
      });
      saveName = '';
      saveDesc = '';
      savePrompt = false;
      await loadFavorites();
    } catch (e) {
      error = (e as Error).message;
    } finally {
      saving = false;
    }
  }

  async function removeFavorite(f: Favorite, ev: Event) {
    ev.stopPropagation();
    if (!confirm(`Delete favorite "${f.name}"?`)) return;
    try {
      await api.deleteFavorite(f.id);
      favorites = favorites.filter((x) => x.id !== f.id);
    } catch (e) {
      error = (e as Error).message;
    }
  }

  let canSave = $derived(currentFlagIDs.length > 0 || currentScriptIDs.length > 0 || currentTargets.trim() !== '');
</script>

<section>
  <div class="head">
    <h2>Profiles &amp; favorites</h2>
    <button onclick={() => (savePrompt = !savePrompt)} disabled={!canSave}>
      {savePrompt ? 'Cancel' : 'Save current as favorite'}
    </button>
  </div>

  {#if savePrompt}
    <div class="save-form">
      <input type="text" placeholder="Name (e.g. 'Home network sweep')" bind:value={saveName} />
      <input type="text" placeholder="Description (optional)" bind:value={saveDesc} />
      <button class="primary" onclick={saveCurrent} disabled={saving || !saveName.trim()}>
        {saving ? 'Saving…' : 'Save'}
      </button>
    </div>
  {/if}

  {#if error}<div class="error">{error}</div>{/if}

  {#if profiles.length > 0}
    <div class="sub">Built-in profiles</div>
    <div class="grid">
      {#each profiles as p (p.id)}
        <button class="card" onclick={() => applyProfile(p)}>
          <div class="row1">
            <span class="icon">{p.icon ?? '•'}</span>
            <span class="name">{p.name}</span>
            {#if p.needs_root}<span class="pill warn">sudo</span>{/if}
            <span class="pill">{p.skill_level}</span>
          </div>
          <p class="desc">{p.description}</p>
        </button>
      {/each}
    </div>
  {/if}

  <div class="sub">Your favorites <span class="count">({favorites.length})</span></div>
  {#if favorites.length === 0}
    <p class="empty">No favorites saved yet. Build a scan in the form above and click "Save current as favorite".</p>
  {/if}
  <div class="grid">
    {#each favorites as f (f.id)}
      <div
        class="card"
        role="button"
        tabindex="0"
        onclick={() => applyFavorite(f)}
        onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); applyFavorite(f); } }}
      >
        <div class="row1">
          <span class="icon">★</span>
          <span class="name">{f.name}</span>
          <button class="del" onclick={(e) => removeFavorite(f, e)} title="Delete favorite">×</button>
        </div>
        {#if f.description}<p class="desc">{f.description}</p>{/if}
        <p class="meta">
          {(f.flag_ids ?? []).length} flags
          {#if (f.script_ids ?? []).length > 0} · {(f.script_ids ?? []).length} scripts{/if}
          {#if f.targets} · {f.targets.length > 30 ? f.targets.slice(0, 30) + '…' : f.targets}{/if}
        </p>
      </div>
    {/each}
  </div>
</section>

<style>
  section {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 1rem 1.1rem;
    box-shadow: var(--shadow);
  }
  .head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 0.6rem; }
  h2 { margin: 0; font-size: 1.05rem; }
  .save-form {
    display: flex;
    gap: 0.5rem;
    align-items: center;
    margin-bottom: 0.85rem;
    flex-wrap: wrap;
  }
  .save-form input { flex: 1 1 200px; }
  .error {
    padding: 0.5rem 0.75rem;
    background: rgba(220, 38, 38, 0.08);
    border: 1px solid rgba(220, 38, 38, 0.2);
    color: var(--danger);
    border-radius: var(--radius);
    margin-bottom: 0.6rem;
  }
  .sub {
    margin: 0.85rem 0 0.4rem;
    color: var(--text-dim);
    font-weight: 600;
    font-size: 0.85rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }
  .count { color: var(--text-dim); text-transform: none; letter-spacing: 0; font-weight: normal; }
  .empty { color: var(--text-dim); font-size: 0.88rem; margin: 0; }
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 0.55rem;
  }
  .card {
    display: flex;
    flex-direction: column;
    text-align: left;
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 0.6rem 0.75rem;
    cursor: pointer;
    gap: 0.25rem;
  }
  .card:hover { border-color: var(--accent); }
  .row1 {
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }
  .icon { font-size: 1.05rem; }
  .name { font-weight: 600; flex: 1; }
  .del {
    background: transparent;
    border: none;
    color: var(--text-dim);
    font-size: 1.1rem;
    line-height: 1;
    padding: 0 0.2rem;
    border-radius: 4px;
  }
  .del:hover { color: var(--danger); background: var(--code-bg); }
  .desc, .meta {
    margin: 0;
    color: var(--text-dim);
    font-size: 0.82rem;
  }
  .meta { font-family: ui-monospace, monospace; font-size: 0.75rem; }
</style>
