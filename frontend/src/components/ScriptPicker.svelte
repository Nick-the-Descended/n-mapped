<script lang="ts">
  import type { NSEScript, SkillLevel } from '../lib/types';

  let {
    scripts,
    selected = $bindable(),
    args = $bindable(),
  }: {
    scripts: NSEScript[];
    selected: Set<string>;
    args: Record<string, string>;
  } = $props();

  let query = $state('');
  let categoryFilter = $state<string>('all');
  let skillFilter = $state<SkillLevel | 'all'>('all');
  let acknowledged = $state<Set<string>>(new Set()); // intrusive scripts the user OK'd

  // Categories present in the catalog (sorted, with safe-first).
  const allCategories = $derived(() => {
    const set = new Set<string>();
    for (const s of scripts) for (const c of s.categories) set.add(c);
    return ['all', ...Array.from(set).sort()];
  });

  const intrusiveCats = new Set(['exploit', 'intrusive', 'dos', 'brute']);

  function isIntrusive(s: NSEScript): boolean {
    return s.categories.some((c) => intrusiveCats.has(c));
  }

  function visible(s: NSEScript): boolean {
    if (categoryFilter !== 'all' && !s.categories.includes(categoryFilter)) return false;
    if (skillFilter !== 'all' && s.skill_level !== skillFilter) return false;
    if (!query.trim()) return true;
    const q = query.toLowerCase();
    return [
      s.id, s.short_description, s.long_description ?? '',
      ...(s.tags ?? []), ...(s.categories ?? []),
    ].some((t) => t.toLowerCase().includes(q));
  }

  function toggle(s: NSEScript) {
    const next = new Set(selected);
    if (next.has(s.id)) {
      next.delete(s.id);
    } else {
      if (isIntrusive(s) && !acknowledged.has(s.id)) {
        const yes = confirm(
          `${s.id} is an intrusive script (categories: ${s.categories.join(', ')}).\n\n` +
            'It can disrupt or actively probe the target. Only run it against systems you have permission to test.\n\nAdd it anyway?',
        );
        if (!yes) return;
        acknowledged = new Set([...acknowledged, s.id]);
      }
      next.add(s.id);
    }
    selected = next;
  }

  function setArg(scriptID: string, argName: string, value: string) {
    const key = `${scriptID}.${argName}`;
    if (value === '') {
      const { [key]: _, ...rest } = args;
      args = rest;
    } else {
      args = { ...args, [key]: value };
    }
  }
</script>

<section>
  <div class="head">
    <h2>NSE scripts</h2>
    <span class="count">
      {selected.size} selected
      {#if selected.size > 0}
        <button class="link" onclick={() => (selected = new Set())}>clear</button>
      {/if}
    </span>
  </div>
  <p class="hint">
    NSE scripts add focused checks on top of port discovery — banner grabs, vuln lookups, OS detail, brute-force.
    Pick the ones relevant to your target. Intrusive scripts (red badge) require an extra confirmation.
  </p>

  <div class="controls">
    <input type="text" placeholder="Search scripts (e.g. 'http', 'ssl', 'vuln')" bind:value={query} />
    <select bind:value={categoryFilter}>
      {#each allCategories() as c (c)}
        <option value={c}>{c}</option>
      {/each}
    </select>
    <div class="skills">
      {#each ['all', 'beginner', 'intermediate', 'advanced'] as level (level)}
        <button
          class:active={skillFilter === level}
          onclick={() => (skillFilter = level as SkillLevel | 'all')}
        >{level}</button>
      {/each}
    </div>
  </div>

  <ul class="list">
    {#each scripts.filter(visible) as s (s.id)}
      {@const checked = selected.has(s.id)}
      {@const intrusive = isIntrusive(s)}
      <li class:checked class:intrusive>
        <label>
          <input type="checkbox" checked={checked} onchange={() => toggle(s)} />
          <span class="id">{s.id}</span>
          <span class="desc">{s.short_description}</span>
          <span class="badges">
            {#each s.categories as c (c)}
              <span class="pill">{c}</span>
            {/each}
            {#if intrusive}<span class="pill danger">intrusive</span>{/if}
          </span>
        </label>
        {#if checked}
          {#if s.long_description}
            <p class="long">{s.long_description}</p>
          {/if}
          {#if s.warnings && s.warnings.length > 0}
            <ul class="warnings">
              {#each s.warnings as w}
                <li class="warn-{w.level}">
                  {w.level === 'danger' ? '⛔' : w.level === 'warn' ? '⚠️' : 'ℹ️'} {w.text}
                </li>
              {/each}
            </ul>
          {/if}
          {#if s.args && s.args.length > 0}
            <div class="args">
              <div class="args-label">Script arguments (optional)</div>
              {#each s.args as a (a.name)}
                {@const key = `${s.id}.${a.name}`}
                <div class="arg-row">
                  <label class="arg-label" for={key}>
                    <code>{a.name}</code>
                    <span class="arg-type">{a.type}</span>
                  </label>
                  <input
                    id={key}
                    type="text"
                    placeholder={a.default ?? a.description ?? ''}
                    value={args[key] ?? ''}
                    oninput={(e) => setArg(s.id, a.name, (e.currentTarget as HTMLInputElement).value)}
                  />
                  {#if a.description}
                    <p class="arg-desc">{a.description}</p>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        {/if}
      </li>
    {/each}
  </ul>
</section>

<style>
  section {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 1rem 1.1rem;
    box-shadow: var(--shadow);
  }
  .head { display: flex; justify-content: space-between; align-items: baseline; }
  h2 { margin: 0; font-size: 1.05rem; }
  .count { color: var(--text-dim); font-size: 0.85rem; }
  .link {
    background: transparent;
    border: none;
    color: var(--accent);
    padding: 0;
    margin-left: 0.4rem;
    cursor: pointer;
    text-decoration: underline;
  }
  .hint { color: var(--text-dim); font-size: 0.88rem; margin: 0.25rem 0 0.6rem; }

  .controls {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
    margin-bottom: 0.65rem;
  }
  .controls input { flex: 1 1 220px; }
  .controls select {
    background: var(--bg-elev);
    color: var(--text);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
    padding: 0.4rem 0.55rem;
  }
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

  .list { list-style: none; margin: 0; padding: 0; }
  .list li {
    border-top: 1px solid var(--border);
    padding: 0.55rem 0.2rem 0.6rem 0.2rem;
  }
  .list li:first-child { border-top: none; }
  .list li.checked { background: rgba(37, 99, 235, 0.05); }
  .list li.intrusive label .id { color: var(--danger); }
  label {
    display: grid;
    grid-template-columns: 1.1rem 14rem 1fr auto;
    gap: 0.5rem;
    align-items: center;
    cursor: pointer;
  }
  .id {
    font-family: ui-monospace, "SF Mono", Menlo, Consolas, monospace;
    font-weight: 600;
    color: var(--accent);
  }
  .badges { display: flex; gap: 0.25rem; flex-wrap: wrap; }
  .long {
    margin: 0.4rem 0 0.2rem 1.6rem;
    color: var(--text-dim);
    font-size: 0.85rem;
    max-width: 56rem;
  }
  ul.warnings { list-style: none; margin: 0.35rem 0 0.2rem 1.6rem; padding: 0; font-size: 0.85rem; }
  ul.warnings li { padding: 0.15rem 0; }
  .warn-warn { color: var(--warn); }
  .warn-danger { color: var(--danger); }
  .warn-info { color: var(--text-dim); }

  .args {
    margin: 0.55rem 0 0.2rem 1.6rem;
    padding: 0.55rem 0.75rem;
    background: var(--code-bg);
    border-radius: var(--radius);
    border: 1px dashed var(--border-strong);
  }
  .args-label {
    font-weight: 600;
    margin-bottom: 0.35rem;
    color: var(--text-dim);
    font-size: 0.85rem;
  }
  .arg-row {
    display: grid;
    grid-template-columns: 16rem 1fr;
    gap: 0.5rem 0.75rem;
    align-items: center;
    margin-top: 0.35rem;
  }
  .arg-label { display: flex; gap: 0.4rem; align-items: baseline; cursor: text; grid-template-columns: none; }
  .arg-type {
    font-size: 0.72rem;
    color: var(--text-dim);
    font-family: ui-monospace, monospace;
  }
  .arg-desc {
    grid-column: 1 / 3;
    margin: 0.05rem 0 0;
    color: var(--text-dim);
    font-size: 0.8rem;
  }
</style>
