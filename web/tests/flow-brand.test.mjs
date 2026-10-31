// node --test tests/: the claim flow order (Sign in with Gully ID > Connect wallet > address > claim), the shared
// Connect-a-wallet dialog filling the address, and brand guards that FAIL on typed wordmarks, kit v3 art or colours,
// and consumer-app links on this Gully Chain surface (owner 2026-10-30).
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, readdirSync, statSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const WEB = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const read = (p) => readFileSync(join(WEB, p), 'utf8');
const faucet = read('src/Faucet.svelte');
const form = faucet.slice(faucet.indexOf('<form id="claim"'), faucet.indexOf('</form>'));

test('form order: Sign in with Gully ID, then Connect wallet, then the address, then the claim', () => {
  const a = form.indexOf('<GullySignIn'), b = form.indexOf('data-testid="connect-wallet"'), c = form.indexOf('id="address"'), d = form.indexOf('data-testid="claim"');
  assert.ok(a > 0 && a < b && b < c && c < d, `order ${a} ${b} ${c} ${d}`);
});

test('stepper copy matches the signin_required flow', () => {
  assert.match(faucet, /needSignIn \? \['Sign in with Gully ID', 'Your address', 'Claim ETH'\]/);
  assert.match(faucet, /cur = needSignIn \? \(paid \? 4 : !gullyToken \? 1 : !\(input \|\| ''\)\.trim\(\) \? 2 : 3\)/);
  assert.match(form, /disabled=\{isLoading \|\| claimBlocked\}/, 'claim disabled until signed in');
  assert.match(form, /data-testid="claim-why"/, 'with a reason');
  assert.match(faucet, /Sign in with Gully ID, then one claim per account/);
});

test('Connect wallet uses the shared dialog and fills the address with accounts[0]', async () => {
  const start = faucet.indexOf('async function connectWallet()');
  const fn = faucet.slice(start, faucet.indexOf('\n  }\n', start) + 4);
  assert.match(fn, /window\.GullyConnect/);
  assert.match(fn, /gc\.open\(\{ appName: 'Gully Faucet' \}\)/);
  const run = new Function('window', 'short', 'notify', `let input = '', step = 1, fieldError = 'x', connecting = false, walletName = '';
    ${fn}
    return connectWallet().then(() => ({ input, step, fieldError, walletName, connecting }));`);
  const ok = { GullyConnect: { open: async (o) => { assert.equal(o.appName, 'Gully Faucet'); return { name: 'Gully Wallet', accounts: ['0xAbC0000000000000000000000000000000000001'] }; } } };
  const r = await run(ok, (a) => a.slice(0, 6), () => {});
  assert.equal(r.input, '0xAbC0000000000000000000000000000000000001');
  assert.equal(r.walletName, 'Gully Wallet');
  assert.equal(r.fieldError, '');
  assert.equal(r.step, 2);
  assert.equal(r.connecting, false);
  const closed = { GullyConnect: { open: async () => { const e = new Error('closed'); e.code = 4001; throw e; } } };
  const c = await run(closed, String, () => assert.fail('no toast when the user closes the dialog'));
  assert.equal(c.input, '');
});

test('gc-connect is loaded on the page and is the canonical 1.1.0 copy (Sign in with Gully ID first)', () => {
  assert.match(read('index.html'), /<script src="\/gc-connect\.js\?v=1\.1\.0" defer><\/script>/);
  assert.match(read('public/gc-connect.js'), /GC_CONNECT_VERSION = '1\.1\.0'/);
});

// ---- brand guards ----
const TEXT = /\.(svelte|js|mjs|css|html|json|webmanifest|svg|txt)$/;
function walk(dir, out = []) {
  for (const n of readdirSync(join(WEB, dir))) {
    const p = dir + '/' + n;
    if (statSync(join(WEB, p)).isDirectory()) walk(p, out);
    else out.push(p);
  }
  return out;
}
const PUBLIC = walk('public');
const OWN = [
  ...walk('src').filter((f) => TEXT.test(f) && !f.startsWith('src/night/')),
  ...PUBLIC.filter((f) => TEXT.test(f) && !/gc-(switcher|connect)\.js$/.test(f)),
  'index.html',
  '404.html',
];

test('no typed GullyChain wordmark: logos are the kit v4 icon image', () => {
  for (const f of OWN) {
    const s = read(f);
    assert.doesNotMatch(s, /class="(logo|brand|fwm|nf-logo|wm)[^"]*"[^>]*>\s*Gully ?Chain\s*</i, `${f}: typed wordmark`);
    assert.doesNotMatch(s, /class="[^"]*\b(wordmark|wm)\b[^"]*"/i, `${f}: wordmark element`);
  }
});

test('no kit v3 art, fonts or colours in the faucet sources and public files', () => {
  const v3 = /gullychain-wordmark|gullychain-logo|gullychain-mascot|gullychain-icon|gullychain-powered-by|wordmark-for-(dark|light)|Exo 2|exo2|Barlow|#(05080D|0A1018|2563FF|00B2FA|46FB04|01D9FA|1567F9)(?![0-9a-f])/i;
  for (const f of OWN) assert.doesNotMatch(read(f), v3, f);
  for (const f of PUBLIC) assert.doesNotMatch(f, /gullychain-(logo|wordmark|mascot|icon)/, `v3 file shipped: ${f}`);
});

test('no consumer-app (kre8) links on this Gully Chain surface; builders go to the quickstart', () => {
  for (const f of OWN) assert.doesNotMatch(read(f), /kre8/i, f);
  assert.match(faucet, /\{ name: 'Start building', href: 'https:\/\/docs\.gullychain\.xyz\/quickstart\/', icon: 'rocket' \}/);
});
