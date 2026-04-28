<script lang="ts">
  import type { TaskProgress } from '../lib/types';

  let { progress, finished }: { progress: TaskProgress | null; finished: boolean } = $props();

  function fmtETA(secs: number): string {
    if (secs < 60) return `${secs}s`;
    const m = Math.floor(secs / 60);
    const s = secs % 60;
    return `${m}m ${s}s`;
  }
</script>

{#if progress || finished}
  <div class="bar" class:done={finished}>
    <div class="track">
      <div class="fill" style:width="{finished ? 100 : Math.max(0, Math.min(100, progress?.percent ?? 0))}%"></div>
    </div>
    <div class="meta">
      {#if finished}
        <span>complete</span>
      {:else if progress}
        <span>{(progress.percent ?? 0).toFixed(1)}%</span>
        {#if progress.task}<span class="task">· {progress.task}</span>{/if}
        {#if progress.remaining && progress.remaining > 0}
          <span>· ~{fmtETA(progress.remaining)} left</span>
        {/if}
      {/if}
    </div>
  </div>
{/if}

<style>
  .bar { margin-top: 0.6rem; }
  .track {
    width: 100%;
    height: 6px;
    background: var(--code-bg);
    border-radius: 3px;
    overflow: hidden;
  }
  .fill {
    height: 100%;
    background: var(--accent);
    transition: width 0.3s ease-out;
  }
  .bar.done .fill { background: var(--ok); }
  .meta {
    display: flex;
    gap: 0.4rem;
    font-size: 0.78rem;
    color: var(--text-dim);
    margin-top: 0.25rem;
  }
  .task { font-style: italic; }
</style>
