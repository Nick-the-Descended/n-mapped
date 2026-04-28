<script lang="ts">
  import type { Catalog, Flag, PrivilegeState, SkillLevel } from '../lib/types';

  let {
    catalog,
    privilege,
    selected = $bindable(),
    values = $bindable(),
  }: {
    catalog: Catalog;
    privilege: PrivilegeState | null;
    selected: Set<string>;
    values: Record<string, string>;
  } = $props();

  let query = $state('');
  let skillFilter = $state<SkillLevel | 'all'>('all');
  let collapsed = $state<Record<string, boolean>>({});

  const allowsRaw = $derived(privilege?.mode === 'root' || privilege?.mode === 'capability');

  function visible(f: Flag): boolean {
    if (skillFilter !== 'all' && f.skill_level !== skillFilter) return false;
    if (!query.trim()) return true;
    const q = query.toLowerCase();
    return [
      f.id, f.short ?? '', f.long ?? '', f.short_description, f.long_description ?? '',
      ...(f.tags ?? []),
    ].some((s) => s.toLowerCase().includes(q));
  }

  function toggle(f: Flag) {
    if (f.requires_root && !allowsRaw) return;
    const next = new Set(selected);
    if (next.has(f.id)) {
      next.delete(f.id);
      const { [f.id]: _, ...rest } = values;
      values = rest;
    } else {
      next.add(f.id);
      // Drop any conflicting selections in the same exclusivity groups.
      for (const group of f.mutually_exclusive_with ?? []) {
        for (const other of catalog.flags) {
          if (other.id === f.id) continue;
          if ((other.mutually_exclusive_with ?? []).includes(group) && next.has(other.id)) {
            next.delete(other.id);
            const { [other.id]: _, ...rest } = values;
            values = rest;
          }
        }
      }
    }
    selected = next;
  }

  function categoryFlags(catId: string): Flag[] {
    return catalog.flags.filter((f) => f.category === catId && visible(f));
  }

  function valuePlaceholder(kind: string): string {
    switch (kind) {
      case 'port-list': return '22,80,443 or 1-1024';
      case 'int':       return '1000';
      case 'ip':        return '192.168.1.1';
      default:          return 'value';
    }
  }
</script>

<section class="picker">
  <div class="controls">
    <input
      type="text"
      placeholder="Search flags (e.g. 'syn', 'udp', 'top')"
      bind:value={query}
    />
    <div class="skills">
      {#each ['all', 'beginner', 'intermediate', 'advanced'] as level (level)}
        <button
          class:active={skillFilter === level}
          onclick={() => (skillFilter = level as SkillLevel | 'all')}
        >{level}</button>
      {/each}
    </div>
  </div>

  {#each catalog.categories as cat (cat.id)}
    {@const flags = categoryFlags(cat.id)}
    {#if flags.length > 0}
      <div class="cat">
        <button
          class="cat-header"
          onclick={() => (collapsed[cat.id] = !collapsed[cat.id])}
          aria-expanded={!collapsed[cat.id]}
        >
          <span class="caret">{collapsed[cat.id] ? '▸' : '▾'}</span>
          <span class="cat-name">{cat.name}</span>
          <span class="cat-count">{flags.length}</span>
        </button>
        {#if cat.description}
          <p class="cat-desc">{cat.description}</p>
        {/if}
        {#if !collapsed[cat.id]}
          <ul class="flags">
            {#each flags as f (f.id)}
              {@const checked = selected.has(f.id)}
              {@const blocked = f.requires_root && !allowsRaw}
              <li class:checked class:blocked>
                <label>
                  <input
                    type="checkbox"
                    checked={checked}
                    disabled={blocked}
                    onchange={() => toggle(f)}
                  />
                  <span class="token">{f.short || f.long}</span>
                  <span class="desc">{f.short_description}</span>
                  <span class="badges">
                    {#if f.requires_root}
                      <span class="pill warn" title={blocked ? 'Run with sudo to enable' : 'Uses raw sockets'}>sudo</span>
                    {/if}
                    {#if f.skill_level === 'advanced'}
                      <span class="pill">advanced</span>
                    {/if}
                  </span>
                </label>
                {#if checked && f.value_type && f.value_type !== 'boolean'}
                  <div class="value">
                    <input
                      type="text"
                      placeholder={valuePlaceholder(f.value_type)}
                      value={values[f.id] ?? ''}
                      oninput={(e) => (values = { ...values, [f.id]: (e.currentTarget as HTMLInputElement).value })}
                    />
                    <span class="value-type">{f.value_type}</span>
                  </div>
                {/if}
                {#if checked && f.long_description}
                  <p class="long">{f.long_description}</p>
                {/if}
              </li>
            {/each}
          </ul>
        {/if}
      </div>
    {/if}
  {/each}
</section>

<style>
  .picker {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
    overflow: hidden;
  }
  .controls {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    padding: 0.85rem 1rem;
    border-bottom: 1px solid var(--border);
    flex-wrap: wrap;
  }
  .controls input { flex: 1 1 220px; }
  .skills { display: flex; gap: 0.25rem; }
  .skills button {
    padding: 0.3rem 0.6rem;
    font-size: 0.82rem;
    text-transform: capitalize;
  }
  .skills button.active {
    background: var(--accent);
    color: var(--accent-fg);
    border-color: var(--accent);
  }
  .cat { border-bottom: 1px solid var(--border); }
  .cat:last-child { border-bottom: none; }
  .cat-header {
    width: 100%;
    text-align: left;
    background: transparent;
    border: none;
    padding: 0.7rem 1rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    border-radius: 0;
  }
  .cat-header:hover { background: var(--code-bg); }
  .caret { width: 1em; color: var(--text-dim); }
  .cat-name { font-weight: 600; }
  .cat-count {
    color: var(--text-dim);
    font-size: 0.82rem;
    margin-left: 0.25rem;
  }
  .cat-desc {
    margin: -0.4rem 1rem 0.4rem 2.4rem;
    color: var(--text-dim);
    font-size: 0.82rem;
  }
  ul.flags {
    list-style: none;
    margin: 0;
    padding: 0 0 0.5rem;
  }
  ul.flags li { padding: 0.4rem 1rem 0.4rem 2.1rem; border-top: 1px dashed var(--border); }
  ul.flags li:first-child { border-top: none; }
  ul.flags li.checked { background: rgba(37,99,235,0.06); }
  ul.flags li.blocked { opacity: 0.55; }
  ul.flags li label {
    display: grid;
    grid-template-columns: 1.1rem 5rem 1fr auto;
    gap: 0.5rem;
    align-items: center;
    cursor: pointer;
  }
  ul.flags li.blocked label { cursor: not-allowed; }
  .token {
    font-family: ui-monospace, "SF Mono", Menlo, Consolas, monospace;
    color: var(--accent);
    font-weight: 600;
  }
  .desc { color: var(--text); }
  .badges { display: flex; gap: 0.3rem; }
  .value {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    margin: 0.45rem 0 0.1rem 1.6rem;
  }
  .value input { flex: 1; max-width: 22rem; }
  .value-type {
    font-size: 0.75rem;
    color: var(--text-dim);
    font-family: ui-monospace, monospace;
  }
  .long {
    margin: 0.45rem 0 0.2rem 1.6rem;
    color: var(--text-dim);
    font-size: 0.85rem;
    max-width: 56rem;
  }
</style>
