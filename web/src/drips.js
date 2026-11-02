// The faucet's real payouts: the faucet account's outgoing transactions from the explorer index (the same public API the
// explorer's own address page reads; no invented or sample numbers anywhere). A source that fails gives an error, never a zero.
const WORKSPACE = 'firebaseUserId=38a1fdfc-c052-4f22-9627-76f3c343a6f4&workspace=Gully%20Chain%20Testnet';
const PAGE = 200;
const MAX_PAGES = 5;
const DAY = 86400e3;

const dayKey = (ms) => new Date(ms).toISOString().slice(0, 10);

/** Fetches the newest pages of the account's history until `days` are covered. Throws when the index does not answer. */
export async function fetchPayouts(fetchFn, explorer, account, days, now = Date.now()) {
  const cutoff = now - days * DAY;
  const items = [];
  let covered = false;
  for (let page = 1; page <= MAX_PAGES; page++) {
    const r = await fetchFn(`${explorer}/api/addresses/${account}/transactions?${WORKSPACE}&page=${page}&itemsPerPage=${PAGE}&order=desc`, { headers: { accept: 'application/json' } });
    if (!r.ok) throw new Error(`explorer index ${r.status}`);
    const d = await r.json();
    const rows = Array.isArray(d?.items) ? d.items : null;
    if (!rows) throw new Error('explorer index answered without a list');
    items.push(...rows);
    if (rows.length < PAGE || Date.parse(rows[rows.length - 1].timestamp) < cutoff) { covered = true; break; }
  }
  return { items, covered };
}

/**
 * Per-day payouts (UTC), oldest first, today last. `n: null` = not covered by the pages read, so it is drawn as no data,
 * not as a quiet day. Only transactions sent BY the faucet count, not the top-ups it receives.
 */
export function bucketPayouts(items, account, days, covered = true, now = Date.now()) {
  const me = (account || '').toLowerCase();
  const out = (items || []).filter((t) => (t.from || '').toLowerCase() === me);
  const oldest = covered ? -Infinity : out.reduce((m, t) => Math.min(m, Date.parse(t.timestamp)), Infinity);
  const bars = [];
  for (let i = days - 1; i >= 0; i--) {
    const k = dayKey(now - i * DAY);
    const complete = !Number.isFinite(oldest) || k > dayKey(oldest);
    bars.push({ day: k, n: complete ? 0 : null, eth: 0 });
  }
  const at = new Map(bars.map((b) => [b.day, b]));
  for (const t of out) {
    const b = at.get(dayKey(Date.parse(t.timestamp)));
    if (b && b.n !== null) { b.n += 1; b.eth += Number(t.value || 0) / 1e18; }
  }
  return bars;
}

/** The newest payouts for the list under the chart. */
export const latestPayouts = (items, account, count = 4) => {
  const me = (account || '').toLowerCase();
  return (items || []).filter((t) => (t.from || '').toLowerCase() === me && Number(t.value || 0) > 0).slice(0, count)
    .map((t) => ({ hash: t.hash, to: t.to, eth: Number(t.value) / 1e18, ts: Date.parse(t.timestamp) / 1000, block: t.blockNumber }));
};

/** The top of the y axis: four whole, round steps above the tallest bar. */
export function niceMax(max) {
  if (!(max > 4)) return 4;
  const mag = 10 ** Math.floor(Math.log10(max / 4));
  const step = [1, 1.5, 2, 2.5, 3, 4, 5, 7.5, 10].map((k) => k * mag).find((v) => Number.isInteger(v) && v * 4 >= max) ?? 10 * mag;
  return step * 4;
}
