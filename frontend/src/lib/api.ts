import type {
  Catalog,
  Favorite,
  HistoryRecord,
  HistorySummary,
  NmapInfo,
  PrivilegeState,
  ScanEvent,
  ScanRequest,
  ScanStartResponse,
} from './types';

async function getJSON<T>(path: string): Promise<T> {
  const res = await fetch(path, { headers: { Accept: 'application/json' } });
  if (!res.ok) throw new Error(`${path}: HTTP ${res.status}`);
  return res.json() as Promise<T>;
}

async function postJSON<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  const text = await res.text();
  let parsed: unknown = null;
  try { parsed = text ? JSON.parse(text) : null; } catch { /* leave null */ }
  if (!res.ok) {
    const msg = (parsed && typeof parsed === 'object' && 'error' in parsed)
      ? String((parsed as { error: unknown }).error)
      : text || `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return parsed as T;
}

async function deleteJSON<T>(path: string): Promise<T> {
  const res = await fetch(path, { method: 'DELETE' });
  const text = await res.text();
  let parsed: unknown = null;
  try { parsed = text ? JSON.parse(text) : null; } catch { /* leave null */ }
  if (!res.ok) {
    const msg = (parsed && typeof parsed === 'object' && 'error' in parsed)
      ? String((parsed as { error: unknown }).error)
      : text || `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return parsed as T;
}

export const api = {
  catalog: () => getJSON<Catalog>('/api/catalog'),
  version: () => getJSON<NmapInfo>('/api/version'),
  privilege: () => getJSON<PrivilegeState>('/api/privilege'),
  startScan: (req: ScanRequest) => postJSON<ScanStartResponse>('/api/scans', req),
  stopScan: (id: string) => postJSON<{ stopped: boolean }>(`/api/scans/${id}/stop`, {}),
  history: (limit = 0) => getJSON<HistorySummary[]>(`/api/history${limit ? `?limit=${limit}` : ''}`),
  historyRecord: (id: string) => getJSON<HistoryRecord>(`/api/history/${id}`),
  historyDelete: (id: string) => deleteJSON<{ deleted: boolean }>(`/api/history/${id}`),
  historyXmlURL: (id: string) => `/api/history/${id}/xml`,
  favorites: () => getJSON<Favorite[]>('/api/favorites'),
  saveFavorite: (fav: Partial<Favorite>) => postJSON<Favorite>('/api/favorites', fav),
  deleteFavorite: (id: string) => deleteJSON<{ deleted: boolean }>(`/api/favorites/${id}`),
};

// streamScan opens an SSE connection for the given scan id and invokes
// onEvent for every event. Returns a function that closes the stream.
export function streamScan(
  id: string,
  onEvent: (ev: ScanEvent) => void,
  onError?: (err: Event) => void,
): () => void {
  const src = new EventSource(`/api/scans/${id}/events`);
  src.onmessage = (msg) => {
    try {
      const ev = JSON.parse(msg.data) as ScanEvent;
      onEvent(ev);
      if (ev.kind === 'done' || ev.kind === 'error') src.close();
    } catch {
      /* ignore malformed line */
    }
  };
  src.onerror = (err) => {
    if (onError) onError(err);
    src.close();
  };
  return () => src.close();
}
