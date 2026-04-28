<script lang="ts">
  // Shown once on first launch (per browser). The acknowledgement is stored
  // under localStorage["n-mapped:ack"] = "v1"; bump the value to re-prompt.
  const ACK_KEY = 'n-mapped:ack';
  const ACK_VAL = 'v1';

  let dismissed = $state(false);
  let agreed = $state(false);

  $effect(() => {
    try {
      if (typeof window !== 'undefined') {
        dismissed = window.localStorage.getItem(ACK_KEY) === ACK_VAL;
      }
    } catch { /* private mode or quota */ }
  });

  function accept() {
    if (!agreed) return;
    try { window.localStorage.setItem(ACK_KEY, ACK_VAL); } catch {}
    dismissed = true;
  }
</script>

{#if !dismissed}
  <div class="overlay" role="dialog" aria-modal="true" aria-labelledby="splash-title">
    <div class="card">
      <h1 id="splash-title">Welcome to n-mapped 👋</h1>
      <p>
        Nmap is a powerful security tool. Used responsibly it helps you understand
        the network you operate; used carelessly it can cause real problems for you
        and the people running the systems you scan.
      </p>
      <ul>
        <li><strong>Only scan systems you own or have explicit permission to scan.</strong>
          Unauthorized port scans are illegal in many jurisdictions and against most ISP terms of service.</li>
        <li>Stick to <code>localhost</code> or your own LAN while you're learning.
          <code>scanme.nmap.org</code> is a public target that the Nmap project explicitly authorizes for testing.</li>
        <li>Aggressive scans (<code>-T5</code>, <code>-A</code>, vuln scripts) can crash fragile services.
          Start gentle.</li>
      </ul>
      <p class="more">Read the project's full guidance: <a href="https://nmap.org/book/legal-issues.html" target="_blank" rel="noopener noreferrer">Legal Issues with Nmap Scanning</a>.</p>
      <label class="ack">
        <input type="checkbox" bind:checked={agreed} />
        I will only scan systems I have permission to scan.
      </label>
      <div class="actions">
        <button class="primary" disabled={!agreed} onclick={accept}>Continue</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    backdrop-filter: blur(2px);
    z-index: 100;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1rem;
  }
  .card {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 10px;
    box-shadow: var(--shadow);
    max-width: 540px;
    width: 100%;
    padding: 1.5rem 1.6rem;
  }
  h1 { margin: 0 0 0.6rem; font-size: 1.25rem; }
  p { margin: 0.4rem 0; }
  ul { padding-left: 1.2rem; margin: 0.4rem 0; }
  ul li { margin: 0.35rem 0; }
  .more { color: var(--text-dim); font-size: 0.88rem; margin-top: 0.5rem; }
  .ack {
    display: flex;
    gap: 0.5rem;
    margin: 0.85rem 0 0.5rem;
    padding: 0.6rem 0.75rem;
    background: var(--code-bg);
    border-radius: var(--radius);
    align-items: flex-start;
  }
  .actions { text-align: right; margin-top: 0.6rem; }
</style>
