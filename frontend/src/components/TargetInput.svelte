<script lang="ts">
  let { value = $bindable('') }: { value: string } = $props();
</script>

<section>
  <label for="targets">Targets</label>
  <p class="hint">
    One or more IPs, CIDR ranges, or hostnames. Examples: <code>scanme.nmap.org</code>,
    <code>192.168.1.0/24</code>, <code>10.0.0.1-50</code>.
    Separate multiple targets with spaces, commas, or newlines.
  </p>
  <textarea
    id="targets"
    bind:value
    rows="2"
    placeholder="scanme.nmap.org"
    spellcheck="false"
  ></textarea>
</section>

<style>
  section {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 1rem 1.1rem;
    box-shadow: var(--shadow);
  }
  label {
    display: block;
    font-weight: 600;
    margin-bottom: 0.25rem;
  }
  .hint {
    margin: 0 0 0.6rem;
    color: var(--text-dim);
    font-size: 0.85rem;
  }
  textarea {
    resize: vertical;
    min-height: 2.4rem;
    font-family: ui-monospace, "SF Mono", Menlo, Consolas, monospace;
  }
</style>
