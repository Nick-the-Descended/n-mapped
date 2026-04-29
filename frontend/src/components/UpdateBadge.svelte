<script lang="ts">
  import { onMount } from 'svelte';
  import { api, type UpdateStatus } from '../lib/api';

  let status = $state<UpdateStatus | null>(null);

  async function refresh() {
    try { status = await api.update(); }
    catch { /* update check is best-effort */ }
  }

  // Initial fetch is delayed slightly so the backend's first check can complete
  // (it's deliberately staggered 5s after boot). We re-poll every 5 minutes.
  onMount(() => {
    const initial = setTimeout(refresh, 6000);
    const interval = setInterval(refresh, 5 * 60 * 1000);
    return () => { clearTimeout(initial); clearInterval(interval); };
  });
</script>

{#if status?.enabled && status.has_update && status.latest_version}
  <a class="badge" href={status.url} target="_blank" rel="noopener noreferrer"
     title="Latest: {status.latest_version} (you're on {status.current_version})">
    🆕 update available — {status.latest_version}
  </a>
{/if}

<style>
  .badge {
    display: inline-block;
    background: rgba(37, 99, 235, 0.12);
    border: 1px solid rgba(37, 99, 235, 0.3);
    color: var(--accent);
    border-radius: 999px;
    padding: 0.1rem 0.6rem;
    font-size: 0.78rem;
    text-decoration: none;
  }
  .badge:hover { background: rgba(37, 99, 235, 0.2); }
</style>
