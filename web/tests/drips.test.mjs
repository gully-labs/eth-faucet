// node --test tests/: the drip chart only draws real payouts from the explorer index, a failed source never reads as zero.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { bucketPayouts, fetchPayouts, latestPayouts } from '../src/drips.js';

const ME = '0x58c33B44938aFE5F133B6FB7C487B60461c0177b';
const NOW = Date.parse('2026-11-02T12:00:00Z');
const tx = (ts, from = ME, value = '3000000000000000', hash = ts) => ({ hash, from, to: '0x1111111111111111111111111111111111111111', value, timestamp: ts, blockNumber: 1 });

test('counts only what the faucet sent, per UTC day, oldest first and today last', () => {
  const items = [tx('2026-11-02T01:00:00Z', ME.toLowerCase()), tx('2026-11-02T00:00:01Z'), tx('2026-11-01T23:59:59Z'), tx('2026-11-01T10:00:00Z', '0x92cfc047652d6aef7eb0fa2744b912d9f627fa41')];
  const b = bucketPayouts(items, ME, 3, true, NOW);
  assert.deepEqual(b.map((x) => [x.day, x.n]), [['2026-10-31', 0], ['2026-11-01', 1], ['2026-11-02', 2]]);
  assert.ok(Math.abs(b[2].eth - 0.006) < 1e-9);
});

test('days the pages read did not reach are no data, not quiet days', () => {
  const b = bucketPayouts([tx('2026-11-02T01:00:00Z'), tx('2026-11-01T05:00:00Z')], ME, 4, false, NOW);
  assert.deepEqual(b.map((x) => x.n), [null, null, null, 1], 'the oldest day read may be partial, so it counts as unread too');
});

test('reads pages until the window is covered, and a failing index throws', async () => {
  const urls = [];
  const full = Array.from({ length: 200 }, () => tx('2026-11-02T01:00:00Z'));
  const ok = async (u) => { urls.push(u); return new Response(JSON.stringify({ items: urls.length === 1 ? full : [tx('2026-10-01T00:00:00Z')] })); };
  const r = await fetchPayouts(ok, 'https://x', ME, 14, NOW);
  assert.equal(urls.length, 2);
  assert.equal(r.covered, true);
  assert.match(urls[0], /\/api\/addresses\/0x58c3.*\/transactions\?.*page=1&itemsPerPage=200&order=desc/);
  await assert.rejects(fetchPayouts(async () => new Response('no', { status: 502 }), 'https://x', ME, 14, NOW), /502/);
  await assert.rejects(fetchPayouts(async () => new Response('{}'), 'https://x', ME, 14, NOW), /without a list/);
});

test('the latest list has outgoing value transfers only, newest first', () => {
  const l = latestPayouts([tx('2026-11-02T02:00:00Z', ME, '0'), tx('2026-11-02T01:00:00Z'), tx('2026-11-01T01:00:00Z', '0x92cfc047652d6aef7eb0fa2744b912d9f627fa41')], ME);
  assert.equal(l.length, 1);
  assert.equal(l[0].eth, 0.003);
});

test('the page has no sample numbers: a dash until data arrives, and the reason when a source fails', () => {
  const s = readFileSync(new URL('../src/Faucet.svelte', import.meta.url), 'utf8');
  assert.match(s, /shownNonce == null \? '–'/);
  assert.match(s, /claimsLeft == null \? '–'/);
  assert.match(s, /data-testid="drips-error"/);
  assert.doesNotMatch(s, /Math\.random/);
  assert.doesNotMatch(s, /payout: 1,/, 'no made-up payout before /api/info answers');
});
