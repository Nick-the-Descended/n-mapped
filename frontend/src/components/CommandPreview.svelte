<script lang="ts">
  let { command, summary }: { command: string; summary: string } = $props();
  let copied = $state(false);

  async function copy() {
    try {
      await navigator.clipboard.writeText(command);
      copied = true;
      setTimeout(() => (copied = false), 1500);
    } catch {
      /* clipboard may be denied; user can still select manually */
    }
  }
</script>

<section>
  <div class="row">
    <div class="label">Command preview</div>
    <button onclick={copy}>{copied ? 'Copied!' : 'Copy'}</button>
  </div>
  <pre><code>{command}</code></pre>
  <div class="label muted">In plain English</div>
  <pre class="summary">{summary}</pre>
</section>

<style>
  section {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 1rem 1.1rem;
    box-shadow: var(--shadow);
  }
  .row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 0.4rem;
  }
  .label { font-weight: 600; }
  .label.muted { color: var(--text-dim); margin-top: 0.85rem; }
  pre { white-space: pre-wrap; word-break: break-word; }
  pre.summary {
    background: transparent;
    padding: 0.25rem 0;
    color: var(--text-dim);
    font-family: inherit;
    font-size: 0.9rem;
  }
</style>
