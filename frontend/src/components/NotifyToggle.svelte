<script lang="ts">
  import { notificationConsent, requestPermission, setNotificationConsent } from '../lib/notify';

  let consent = $state(notificationConsent());
  let busy = $state(false);
  let supported = $state(true);

  $effect(() => {
    if (typeof Notification === 'undefined') supported = false;
  });

  async function enable() {
    busy = true;
    const result = await requestPermission();
    busy = false;
    if (result === 'granted') {
      consent = 'yes';
      setNotificationConsent('yes');
    } else if (result === 'denied' || result === 'unsupported') {
      consent = 'no';
      setNotificationConsent('no');
    }
  }

  function disable() {
    consent = 'no';
    setNotificationConsent('no');
  }
</script>

{#if !supported}
  <span class="pill">notifications unsupported</span>
{:else if consent === 'yes'}
  <button onclick={disable}>Mute notifications</button>
{:else}
  <button onclick={enable} disabled={busy}>
    {busy ? '…' : 'Notify on scan finish'}
  </button>
{/if}

<style>
  button { font-size: 0.82rem; padding: 0.3rem 0.6rem; }
</style>
