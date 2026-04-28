// Thin wrapper around the browser Notifications API. Requests permission
// lazily on first use and silently no-ops when permission is denied.
//
// Stored consent: localStorage["n-mapped:notify"] = "yes" | "no" | "unset"

const KEY = 'n-mapped:notify';
type Consent = 'yes' | 'no' | 'unset';

function readConsent(): Consent {
  try {
    const v = localStorage.getItem(KEY);
    if (v === 'yes' || v === 'no') return v;
  } catch { /* private mode */ }
  return 'unset';
}

function writeConsent(c: Consent) {
  try { localStorage.setItem(KEY, c); } catch {}
}

export function notificationConsent(): Consent {
  return readConsent();
}

export function setNotificationConsent(c: Consent) {
  writeConsent(c);
}

export async function requestPermission(): Promise<NotificationPermission | 'unsupported'> {
  if (typeof Notification === 'undefined') return 'unsupported';
  if (Notification.permission === 'granted') return 'granted';
  if (Notification.permission === 'denied') return 'denied';
  return Notification.requestPermission();
}

export function notify(title: string, body?: string) {
  if (readConsent() !== 'yes') return;
  if (typeof Notification === 'undefined') return;
  if (Notification.permission !== 'granted') return;
  try {
    new Notification(title, { body, silent: false });
  } catch {
    /* permission revoked or transient failure */
  }
}
