<script lang="ts">
  import { api } from '../lib/api';

  let {
    historyID,
    tags = $bindable<string[]>([]),
    notes = $bindable<string>(''),
  }: {
    historyID: string;
    tags: string[];
    notes: string;
  } = $props();

  let draftTag = $state('');
  let saving = $state(false);
  let savedAt = $state<number | null>(null);
  let error = $state<string | null>(null);
  let savingTimer: ReturnType<typeof setTimeout> | null = null;

  function scheduleSave() {
    if (savingTimer) clearTimeout(savingTimer);
    savingTimer = setTimeout(persist, 600);
  }

  async function persist() {
    saving = true;
    error = null;
    try {
      const updated = await api.historyMeta(historyID, tags, notes);
      tags = updated.tags ?? [];
      notes = updated.notes ?? '';
      savedAt = Date.now();
    } catch (e) {
      error = (e as Error).message;
    } finally {
      saving = false;
    }
  }

  function addTag() {
    const t = draftTag.trim();
    if (!t) return;
    if (tags.includes(t)) { draftTag = ''; return; }
    if (tags.length >= 32) {
      error = 'Max 32 tags per scan.';
      return;
    }
    if (t.length > 64) {
      error = 'Tag too long (max 64 chars).';
      return;
    }
    tags = [...tags, t];
    draftTag = '';
    scheduleSave();
  }

  function removeTag(t: string) {
    tags = tags.filter((x) => x !== t);
    scheduleSave();
  }

  function onTagKeydown(ev: KeyboardEvent) {
    if (ev.key === 'Enter' || ev.key === ',') {
      ev.preventDefault();
      addTag();
    }
  }

  function onNotesInput() { scheduleSave(); }
</script>

<section>
  <div class="row">
    <span class="lbl">Tags</span>
    <div class="tags">
      {#each tags as t (t)}
        <span class="tag">
          {t}
          <button onclick={() => removeTag(t)} aria-label="Remove tag {t}">×</button>
        </span>
      {/each}
      <input
        type="text"
        placeholder="add tag, ↵ to confirm"
        bind:value={draftTag}
        onkeydown={onTagKeydown}
        onblur={addTag}
      />
    </div>
  </div>
  <div class="row">
    <span class="lbl">Notes</span>
    <textarea rows="2" placeholder="Free-form notes about this scan…" bind:value={notes} oninput={onNotesInput}></textarea>
  </div>
  <div class="status">
    {#if error}
      <span class="err">⚠ {error}</span>
    {:else if saving}
      <span>saving…</span>
    {:else if savedAt}
      <span>saved · {new Date(savedAt).toLocaleTimeString()}</span>
    {:else}
      <span class="muted">edits auto-save 0.6s after typing stops</span>
    {/if}
  </div>
</section>

<style>
  section {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 0.85rem 1rem;
    box-shadow: var(--shadow);
  }
  .row { display: grid; grid-template-columns: 5rem 1fr; gap: 0.7rem; align-items: start; margin-bottom: 0.55rem; }
  .lbl { padding-top: 0.4rem; color: var(--text-dim); font-weight: 600; font-size: 0.85rem; }
  .tags { display: flex; flex-wrap: wrap; gap: 0.3rem; align-items: center; }
  .tag {
    display: inline-flex; align-items: center; gap: 0.3rem;
    padding: 0.15rem 0.55rem;
    background: var(--code-bg);
    border: 1px solid var(--border-strong);
    border-radius: 999px;
    font-size: 0.85rem;
  }
  .tag button {
    background: transparent; border: none; padding: 0; line-height: 1;
    color: var(--text-dim); cursor: pointer; font-size: 1rem;
  }
  .tag button:hover { color: var(--danger); }
  .tags input {
    flex: 1 1 8rem;
    min-width: 8rem;
  }
  textarea { resize: vertical; min-height: 2.4rem; }
  .status { font-size: 0.78rem; color: var(--text-dim); margin-top: 0.3rem; min-height: 1em; }
  .status .err { color: var(--danger); }
  .muted { color: var(--text-dim); }
</style>
