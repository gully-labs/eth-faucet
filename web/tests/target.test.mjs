// node --test tests/: who a claim pays (owner report 2026-10-31: a claim paid the Gully ID main wallet while kre8 acts
// as the kre8 passkey account). Pure checks of src/target.js, then Faucet.svelte compiled with svelte/compiler and
// mounted in happy-dom: an app link (?to=…&from=kre8) locks the field, and a signed-in Gully ID user who types another
// address gets "This is not your kre8 / Gully ID account" with a one-tap "Use my Gully account".
import { test, after } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, writeFileSync, mkdtempSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { Window } from 'happy-dom';
import { compile } from 'svelte/compiler';
import { parseTarget, targetCheck, normAddress, sameAddress } from '../src/target.js';

const KRE8 = '0x1111111111111111111111111111111111111111';
const MAIN = '0x2222222222222222222222222222222222222222';
const LINKED = '0x3333333333333333333333333333333333333333';
const OTHER = '0x4444444444444444444444444444444444444444';

// ------------------------------------------------------------------ pure

test('parseTarget: ?to= (and the old ?address=) from an app link, checksummed; unknown apps ignored', () => {
  assert.deepEqual(parseTarget(`?to=${KRE8}&from=kre8`), { to: KRE8, from: 'kre8' });
  assert.deepEqual(parseTarget(`?address=${KRE8}`), { to: KRE8, from: null });
  assert.deepEqual(parseTarget(`?to=${KRE8}&from=evil`), { to: KRE8, from: null });
  assert.deepEqual(parseTarget('?to=0x123&from=kre8'), { to: null, from: null });
  assert.deepEqual(parseTarget(''), { to: null, from: null });
  const mixed = '0xAbCdEf0123456789aBCdef0123456789AbcDEF01';
  assert.equal(parseTarget(`?to=${mixed.toLowerCase()}`).to, normAddress(mixed));
});

test('targetCheck: nothing to compare until signed in with a Gully account and a full address is in the field', () => {
  assert.equal(targetCheck({ field: OTHER, signedIn: false, gullyMain: MAIN }).state, 'idle');
  assert.equal(targetCheck({ field: 'vitalik.eth', signedIn: true, gullyMain: MAIN }).state, 'idle');
  assert.equal(targetCheck({ field: '', signedIn: true, gullyMain: MAIN }).state, 'idle');
  assert.equal(targetCheck({ field: OTHER, signedIn: true, gullyMain: null }).state, 'idle');
});

test('targetCheck: the Gully ID main wallet is fine; any other address is "not your kre8 / Gully ID account"', () => {
  assert.equal(targetCheck({ field: MAIN.toUpperCase().replace('0X', '0x'), signedIn: true, gullyMain: MAIN }).state, 'ok');
  const c = targetCheck({ field: OTHER, signedIn: true, gullyMain: MAIN, gullyWallets: [{ address: MAIN }] });
  assert.equal(c.state, 'mismatch');
  assert.equal(c.title, 'This is not your kre8 / Gully ID account');
  assert.equal(c.expected, MAIN);
  assert.equal(c.own, false);
  const own = targetCheck({ field: LINKED, signedIn: true, gullyMain: MAIN, gullyWallets: [{ address: MAIN }, { address: LINKED }] });
  assert.equal(own.state, 'mismatch');
  assert.equal(own.own, true);
  assert.match(own.detail, /one of your Gully ID wallets, but kre8 does not use it/);
});

test('targetCheck: with an app link the expected account is the app account, not the main wallet', () => {
  assert.equal(targetCheck({ field: KRE8, signedIn: true, gullyMain: MAIN, lockedTo: KRE8, from: 'kre8' }).state, 'ok');
  const c = targetCheck({ field: MAIN, signedIn: true, gullyMain: MAIN, gullyWallets: [{ address: MAIN }], lockedTo: KRE8, from: 'kre8' });
  assert.equal(c.state, 'mismatch');
  assert.equal(c.expected, KRE8);
  assert.match(c.detail, /your kre8 account/);
  assert.equal(sameAddress(KRE8, KRE8.toLowerCase()), true);
  assert.equal(sameAddress('0x12', '0x12'), false);
});

// ------------------------------------------------------------------ the page

const WEB = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const win = new Window({ url: 'https://faucet.gullychain.xyz/' });
for (const k of ['window', 'document', 'navigator', 'HTMLElement', 'Element', 'Node', 'Text', 'Comment', 'DocumentFragment', 'MutationObserver', 'getComputedStyle', 'requestAnimationFrame', 'cancelAnimationFrame', 'Event', 'KeyboardEvent', 'location']) {
  if (!(k in globalThis) || ['window', 'document', 'location', 'Event', 'KeyboardEvent'].includes(k)) Object.defineProperty(globalThis, k, { value: k === 'window' ? win : win[k], configurable: true, writable: true });
}

let signedIn = null; // the fake Gully ID session the SDK reports
function fakeSdk() {
  return {
    version: '1.2.0',
    init: () => ({
      handleRedirect: async () => null,
      session: () => signedIn,
      onChange: () => () => {},
      renderButton(el) {
        const b = win.document.createElement('button');
        b.textContent = 'Sign in with Gully ID';
        el.appendChild(b);
        return () => b.remove();
      },
    }),
  };
}
const INFO = { account: '0x58c33B44938aFE5F133B6FB7C487B60461c0177b', network: 'GullyChain Testnet', payout: '0.0005', symbol: 'ETH', hcaptcha_sitekey: '', gullyid_client: 'gully-faucet', gullyid_issuer: 'https://gullyid.xyz', signin_required: true, eligibility_check: true, window_hours: 24 };
const claims = [];
globalThis.fetch = async (url, init) => {
  const json = (v, status = 200) => new Response(JSON.stringify(v), { status, headers: { 'content-type': 'application/json' } });
  if (url === '/api/info') return json(INFO);
  if (url === '/api/claim') {
    claims.push(JSON.parse(init.body));
    return json({ msg: `Txhash: 0x${'ab'.repeat(32)}` });
  }
  return json({ jsonrpc: '2.0', id: 1, result: '0x0' });
};
win.fetch = globalThis.fetch;

const dir = mkdtempSync(join(WEB, 'node_modules', '.target-test-'));
const comp = (name) => {
  let { js } = compile(readFileSync(join(WEB, 'src', name), 'utf8'), { generate: 'client', filename: name, css: 'injected' });
  writeFileSync(join(dir, name.replace(/\.svelte$/, '.js')), js.code.replace(/'\.\/GullySignIn\.svelte'/g, "'./GullySignIn.js'"));
};
comp('GullySignIn.svelte');
comp('Faucet.svelte');
for (const f of ['gullyid.js', 'target.js']) writeFileSync(join(dir, f), readFileSync(join(WEB, 'src', f)));
const { mount, unmount, flushSync, tick } = await import('svelte');
const { default: Faucet } = await import(pathToFileURL(join(dir, 'Faucet.js')).href);

const mounted = [];
after(() => mounted.forEach((m) => unmount(m)));
async function page(url) {
  win.happyDOM.setURL(url);
  win.document.body.innerHTML = '';
  const target = win.document.createElement('div');
  win.document.body.appendChild(target);
  mounted.push(mount(Faucet, { target }));
  await settle();
  return target;
}
async function settle(n = 30) {
  for (let i = 0; i < n; i++) {
    await new Promise((r) => setTimeout(r, 4));
    flushSync();
  }
  await tick();
}
const typeInto = async (input, v) => {
  input.value = v;
  input.dispatchEvent(new win.Event('input', { bubbles: true }));
  await settle(5);
};

test('an app link locks the field to the kre8 account, and signing in does not swap it for the main wallet', async () => {
  win.GullyID = fakeSdk();
  signedIn = { accessToken: 'gid_test_token_1234', expiresAt: Date.now() + 3600e3, user: { name: 'owner', wallet: { address: MAIN }, wallets: [{ address: MAIN, kind: 'kernel-v3.3' }] } };
  const t = await page(`https://faucet.gullychain.xyz/?to=${KRE8.toLowerCase()}&from=kre8`);
  const input = t.querySelector('[data-testid="address"]');
  assert.equal(input.value, KRE8);
  assert.equal(input.readOnly, true);
  assert.match(t.querySelector('[data-testid="locked-to"]').textContent, /Pays your kre8 account 0x1111…1111/);
  assert.equal(t.querySelector('[data-testid="not-your-account"]'), null, 'the kre8 account is the right one: no warning');
  assert.match(t.querySelector('.gid-who')?.textContent || '', /the claim pays your kre8 account/);
  t.querySelector('form').dispatchEvent(new win.Event('submit', { bubbles: true, cancelable: true }));
  await settle();
  assert.deepEqual(claims.at(-1), { address: KRE8 }, 'the claim pays exactly the kre8 account');
});

test('signed in, another address in the field: warning first, "Use my Gully account" fixes it in one tap', async () => {
  claims.length = 0;
  signedIn = { accessToken: 'gid_test_token_1234', expiresAt: Date.now() + 3600e3, user: { name: 'owner', wallet: { address: MAIN }, wallets: [{ address: MAIN, kind: 'kernel-v3.3' }] } };
  const t = await page('https://faucet.gullychain.xyz/');
  const input = t.querySelector('[data-testid="address"]');
  assert.equal(input.value, MAIN, 'signed in: the main wallet is filled in');
  assert.equal(input.readOnly, false);
  await typeInto(input, OTHER);
  const warn = t.querySelector('[data-testid="not-your-account"]');
  assert.ok(warn, 'warned before paying');
  assert.match(warn.textContent, /This is not your kre8 \/ Gully ID account/);
  // first press only warns (nothing paid), the button now says so
  t.querySelector('form').dispatchEvent(new win.Event('submit', { bubbles: true, cancelable: true }));
  await settle();
  assert.equal(claims.length, 0);
  assert.match(t.querySelector('button[type="submit"]').textContent, /Send to this address anyway/);
  t.querySelector('[data-testid="use-my-gully-account"]').click();
  await settle(5);
  assert.equal(input.value, MAIN);
  assert.equal(t.querySelector('[data-testid="not-your-account"]'), null);
  t.querySelector('form').dispatchEvent(new win.Event('submit', { bubbles: true, cancelable: true }));
  await settle();
  assert.deepEqual(claims.at(-1), { address: MAIN });
});
