<script lang="ts">
  import { saveTheme, type Theme } from '../lib/theme';
  import { clearDraft } from '../lib/draft';
  import { api } from '../lib/api';

  let { theme = $bindable<Theme>('system'), onClose }: { theme: Theme; onClose: () => void } = $props();

  function setTheme(t: Theme) {
    theme = t;
    saveTheme(t);
  }

  function clearLocalState() {
    if (!confirm('Clear all locally-stored settings (theme, ethical-use ack, draft, notification consent)?')) return;
    try {
      localStorage.removeItem('n-mapped:theme');
      localStorage.removeItem('n-mapped:ack');
      localStorage.removeItem('n-mapped:notify');
      clearDraft();
    } catch {}
    location.reload();
  }
</script>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<div class="overlay" role="dialog" aria-modal="true" aria-labelledby="settings-title" tabindex="-1" onclick={onClose}>
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="card" role="document" onclick={(e) => e.stopPropagation()}>
    <div class="head">
      <h2 id="settings-title">Settings</h2>
      <button onclick={onClose} aria-label="Close settings">✕</button>
    </div>

    <section>
      <div class="label">Theme</div>
      <div class="seg">
        {#each ['light', 'dark', 'system'] as t}
          <button class:active={theme === t} onclick={() => setTheme(t as Theme)}>{t}</button>
        {/each}
      </div>
      <p class="hint">"System" follows your OS dark/light preference.</p>
    </section>

    <section>
      <div class="label">Local data</div>
      <p class="hint">
        Settings, dismissed splash, and the in-progress builder draft live in your browser's localStorage.
        Scan history and favorites live in <code>~/.local/share/n-mapped/</code> on Linux,
        <code>~/Library/Application Support/n-mapped/</code> on macOS — clear those by deleting the directory.
      </p>
      <button onclick={clearLocalState}>Clear browser-side state</button>
    </section>

    <section>
      <div class="label">Audit log export</div>
      <p class="hint">
        Download every scan you've run as a flat file — useful for compliance reviews, sharing with a team, or feeding a SIEM. Includes timestamps, command, targets, hosts up, open-port count, tags, and notes.
      </p>
      <div class="dl">
        <a class="btn" href={api.auditURL('json')} download>Export JSON</a>
        <a class="btn" href={api.auditURL('csv')} download>Export CSV</a>
      </div>
    </section>

    <section>
      <div class="label">Keyboard shortcuts</div>
      <table class="kbd">
        <tbody>
          <tr><td><kbd>?</kbd></td><td>Show keyboard shortcuts</td></tr>
          <tr><td><kbd>⌘</kbd>+<kbd>K</kbd> / <kbd>Ctrl</kbd>+<kbd>K</kbd></td><td>Focus the flag search</td></tr>
          <tr><td><kbd>⌘</kbd>+<kbd>Enter</kbd> / <kbd>Ctrl</kbd>+<kbd>Enter</kbd></td><td>Run the current scan</td></tr>
          <tr><td><kbd>⌘</kbd>+<kbd>.</kbd> / <kbd>Ctrl</kbd>+<kbd>.</kbd></td><td>Stop the running scan</td></tr>
          <tr><td><kbd>⌘</kbd>+<kbd>S</kbd> / <kbd>Ctrl</kbd>+<kbd>S</kbd></td><td>Save current command as a favorite</td></tr>
          <tr><td><kbd>1</kbd> / <kbd>2</kbd> / <kbd>3</kbd></td><td>Switch to Builder / Results / History tab</td></tr>
          <tr><td><kbd>Esc</kbd></td><td>Close this dialog</td></tr>
        </tbody>
      </table>
    </section>
  </div>
</div>

<style>
  .overlay {
    position: fixed; inset: 0;
    background: rgba(0, 0, 0, 0.45);
    backdrop-filter: blur(2px);
    z-index: 90;
    display: flex; align-items: center; justify-content: center;
    padding: 1rem;
  }
  .card {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 10px;
    box-shadow: var(--shadow);
    max-width: 540px; width: 100%;
    padding: 1.2rem 1.4rem 1rem;
    max-height: 85vh; overflow: auto;
  }
  .head {
    display: flex; justify-content: space-between; align-items: center;
    margin-bottom: 0.8rem;
  }
  h2 { margin: 0; font-size: 1.15rem; }
  .head button {
    background: transparent; border: none; font-size: 1.1rem;
    color: var(--text-dim); padding: 0.2rem 0.5rem;
  }
  section { padding: 0.6rem 0; border-top: 1px solid var(--border); }
  section:first-of-type { border-top: none; padding-top: 0; }
  .label { font-weight: 600; margin-bottom: 0.35rem; }
  .hint { color: var(--text-dim); font-size: 0.85rem; margin: 0.3rem 0 0.5rem; }
  .seg { display: inline-flex; gap: 0.25rem; }
  .seg button { text-transform: capitalize; }
  .seg button.active {
    background: var(--accent); color: var(--accent-fg); border-color: var(--accent);
  }
  table.kbd { width: 100%; font-size: 0.88rem; }
  table.kbd td { padding: 0.2rem 0.4rem; border-bottom: 1px dashed var(--border); }
  table.kbd td:first-child { white-space: nowrap; }
  kbd {
    background: var(--code-bg);
    border: 1px solid var(--border-strong);
    border-bottom-width: 2px;
    border-radius: 4px;
    padding: 0.05em 0.4em;
    font-family: ui-monospace, monospace;
    font-size: 0.85em;
  }
  .dl { display: flex; gap: 0.4rem; }
  .dl .btn {
    display: inline-block;
    padding: 0.4rem 0.75rem;
    background: var(--bg-elev);
    color: var(--text);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
    text-decoration: none;
    font: inherit;
  }
  .dl .btn:hover { border-color: var(--accent); color: var(--accent); }
</style>
