<script lang="ts">
  let { onClose }: { onClose: () => void } = $props();
</script>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<div class="overlay" role="dialog" aria-modal="true" tabindex="-1" onclick={onClose}>
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div class="card" role="document" onclick={(e) => e.stopPropagation()}>
    <div class="head">
      <h2>Keyboard shortcuts</h2>
      <button onclick={onClose} aria-label="Close">✕</button>
    </div>
    <table>
      <tbody>
        <tr><td><kbd>?</kbd></td><td>Show this list</td></tr>
        <tr><td><kbd>⌘ K</kbd> / <kbd>Ctrl K</kbd></td><td>Focus the flag search box</td></tr>
        <tr><td><kbd>⌘ ⏎</kbd> / <kbd>Ctrl ⏎</kbd></td><td>Run scan</td></tr>
        <tr><td><kbd>⌘ .</kbd> / <kbd>Ctrl .</kbd></td><td>Stop running scan</td></tr>
        <tr><td><kbd>⌘ S</kbd> / <kbd>Ctrl S</kbd></td><td>Save as favorite</td></tr>
        <tr><td><kbd>1</kbd> / <kbd>2</kbd> / <kbd>3</kbd></td><td>Switch to Builder / Results / History</td></tr>
        <tr><td><kbd>Esc</kbd></td><td>Close any open dialog</td></tr>
      </tbody>
    </table>
  </div>
</div>

<style>
  .overlay {
    position: fixed; inset: 0;
    background: rgba(0, 0, 0, 0.45);
    backdrop-filter: blur(2px);
    z-index: 95;
    display: flex; align-items: center; justify-content: center;
    padding: 1rem;
  }
  .card {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 10px;
    box-shadow: var(--shadow);
    max-width: 460px; width: 100%;
    padding: 1.1rem 1.3rem;
  }
  .head {
    display: flex; justify-content: space-between; align-items: center;
    margin-bottom: 0.6rem;
  }
  h2 { margin: 0; font-size: 1.05rem; }
  .head button {
    background: transparent; border: none; font-size: 1.05rem;
    color: var(--text-dim); padding: 0.15rem 0.4rem;
  }
  table { width: 100%; font-size: 0.88rem; }
  td { padding: 0.25rem 0.4rem; border-bottom: 1px dashed var(--border); }
  td:first-child { white-space: nowrap; min-width: 9rem; }
  kbd {
    background: var(--code-bg);
    border: 1px solid var(--border-strong);
    border-bottom-width: 2px;
    border-radius: 4px;
    padding: 0.05em 0.4em;
    font-family: ui-monospace, monospace;
    font-size: 0.85em;
  }
</style>
