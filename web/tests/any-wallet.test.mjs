// node --test tests/: any wallet, Gully ID optional (owner 2026-11-05: "i should be able to send test eth to any wallet
// not just gullyid it needs to be option", with the page stuck on "Opening Gully ID…").
// Faucet.svelte + GullySignIn.svelte compiled with svelte/compiler and mounted in happy-dom, /api/info and /api/claim
// faked per test:
//   - optional mode: the address field is first and usable at once, the claim is not locked, "Sign in with Gully ID" is a
//     secondary, clearly optional action under the claim button, and nothing opens Gully ID by itself;
//   - a claim without a session sends no Authorization header; with a session it sends the bearer and any address pays
//     on one press;
//   - a "no history" refusal says how to fund a brand-new wallet and turns the sign-in into the primary button;
//   - "Opening Gully ID…" never sticks (Cancel, and a watchdog);
//   - required mode is as before (claim locked until signed in).
import { test, after } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, writeFileSync, mkdtempSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { Window } from 'happy-dom';
import { compile } from 'svelte/compiler';

const MAIN = '0x2222222222222222222222222222222222222222';
const OTHER = '0x4444444444444444444444444444444444444444';
const FRESH = '0x5555555555555555555555555555555555555555';
const FRESH_MSG = 'This address has no history yet. Sign in with Gully ID to fund a brand-new wallet, or use a wallet that has made at least one transaction on Gully Chain or Arbitrum Sepolia.';

const WEB = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const win = new Window({ url: 'https://faucet.gullychain.xyz/' });
const FORCE = new Set(['window', 'document', 'location', 'customElements', 'localStorage', 'sessionStorage', 'navigator']);
for (const k of new Set([...Object.getOwnPropertyNames(win), ...Object.getOwnPropertyNames(Object.getPrototypeOf(win))])) {
  if (k === 'constructor' || (k in globalThis && !FORCE.has(k))) continue;
  let v;
  try { v = k === 'window' ? win : win[k]; } catch { continue; }
  if (v === undefined) continue;
  if (typeof v === 'function' && /^[a-z]/.test(k)) v = v.bind(win);
  try { Object.defineProperty(globalThis, k, { value: v, configurable: true, writable: true }); } catch {}
}

// ---- the fake Gully ID SDK: counts every call that would open Gully ID
let signedIn = null; // the session the SDK already holds on load
let signInImpl = async () => null; // what a press on the button does
const sdkCalls = { signIn: 0, handleRedirect: 0 };
let onChange = () => {};
function fakeSdk() {
  return {
    version: '1.2.0',
    init: () => ({
      handleRedirect: async () => { sdkCalls.handleRedirect++; return null; },
      session: () => signedIn,
      onChange: (fn) => { onChange = fn; return () => {}; },
      signIn: (o) => { sdkCalls.signIn++; return signInImpl(o); },
      signOut: async () => { signedIn = null; },
    }),
  };
}
const session = () => ({ accessToken: 'gid_test_token_1234', expiresAt: Date.now() + 3600e3, user: { name: 'owner', wallet: { address: MAIN }, wallets: [{ address: MAIN, kind: 'kernel-v3.3' }] } });

// ---- the fake faucet API
const BASE = { account: '0x58c33B44938aFE5F133B6FB7C487B60461c0177b', network: 'GullyChain Testnet', payout: '0.003', symbol: 'ETH', hcaptcha_sitekey: '', gullyid_client: 'gully-faucet', gullyid_issuer: 'https://gullyid.xyz', eligibility_check: true, window_hours: 24 };
const OPTIONAL = { ...BASE, signin_required: false, signin_mode: 'optional' };
const REQUIRED = { ...BASE, signin_required: true, signin_mode: 'required' };
let INFO = OPTIONAL;
let claimAnswer = () => ({ status: 200, body: { msg: `Txhash: 0x${'ab'.repeat(32)}` } });
const claims = []; // { address, auth }
globalThis.fetch = async (url, init) => {
  const json = (v, status = 200) => new Response(JSON.stringify(v), { status, headers: { 'content-type': 'application/json' } });
  if (url === '/api/info') return json(INFO);
  if (url === '/api/claim') {
    const h = new Headers(init.headers);
    const c = { address: JSON.parse(init.body).address, auth: h.get('authorization') };
    claims.push(c);
    const a = claimAnswer(c);
    return json(a.body, a.status);
  }
  return json({ jsonrpc: '2.0', id: 1, result: '0x0' });
};
win.fetch = globalThis.fetch;

const dir = mkdtempSync(join(WEB, 'node_modules', '.anywallet-test-'));
const comp = (name) => {
  const { js } = compile(readFileSync(join(WEB, 'src', name), 'utf8'), { generate: 'client', filename: name, css: 'injected' });
  writeFileSync(join(dir, name.replace(/\.svelte$/, '.js')), js.code.replace(/'\.\/GullySignIn\.svelte'/g, "'./GullySignIn.js'"));
};
comp('GullySignIn.svelte');
comp('Faucet.svelte');
for (const f of ['gullyid.js', 'target.js', 'drips.js']) writeFileSync(join(dir, f), readFileSync(join(WEB, 'src', f)));
const { mount, unmount, flushSync, tick } = await import('svelte');
const { default: Faucet } = await import(pathToFileURL(join(dir, 'Faucet.js')).href);
const { default: GullySignIn } = await import(pathToFileURL(join(dir, 'GullySignIn.js')).href);

const mounted = [];
after(() => mounted.forEach((m) => unmount(m)));
async function settle(n = 30) {
  for (let i = 0; i < n; i++) {
    await new Promise((r) => setTimeout(r, 4));
    flushSync();
  }
  await tick();
}
async function page({ info = OPTIONAL, sess = null, url = 'https://faucet.gullychain.xyz/' } = {}) {
  INFO = info;
  signedIn = sess;
  signInImpl = async () => null;
  claimAnswer = () => ({ status: 200, body: { msg: `Txhash: 0x${'ab'.repeat(32)}` } });
  claims.length = 0;
  sdkCalls.signIn = 0;
  sdkCalls.handleRedirect = 0;
  win.GullyID = fakeSdk();
  win.happyDOM.setURL(url);
  win.document.body.innerHTML = '';
  const target = win.document.createElement('div');
  win.document.body.appendChild(target);
  mounted.push(mount(Faucet, { target }));
  await settle();
  return target;
}
const typeInto = async (input, v) => {
  input.value = v;
  input.dispatchEvent(new win.Event('input', { bubbles: true }));
  await settle(5);
};
const submit = async (t) => {
  t.querySelector('form').dispatchEvent(new win.Event('submit', { bubbles: true, cancelable: true }));
  await settle();
};
const $ = (t, id) => t.querySelector(`[data-testid="${id}"]`);
const before = (a, b) => !!(a.compareDocumentPosition(b) & win.Node.DOCUMENT_POSITION_FOLLOWING);

test('optional mode, signed out: the address field is first and usable at once, sign-in is a secondary optional action', async () => {
  const t = await page();
  const input = $(t, 'address');
  assert.ok(input, 'the address field is on the page without signing in');
  assert.equal(input.readOnly, false);
  assert.equal(input.disabled, false);
  const claim = $(t, 'claim');
  assert.equal(claim.disabled, false, 'the claim button is not locked behind a sign-in');
  assert.match(claim.textContent, /Request 0\.003 ETH/);
  assert.equal($(t, 'claim-why'), null, 'no "unlocks after you sign in" line');

  // the stepper and the copy say what the page does now
  const steps = [...$(t, 'stepper').querySelectorAll('li')].map((li) => li.textContent.replace(/^\d+/, '').trim());
  assert.deepEqual(steps, ['Enter address', 'Request ETH', 'Done']);
  assert.match($(t, 'promise').textContent, /Paste any wallet address/);
  assert.match($(t, 'promise').textContent, /Signing in with Gully ID is optional/);
  assert.doesNotMatch(t.textContent, /Sign in with Gully ID, then one claim per account/);
  assert.match($(t, 'who-is-paid').textContent, /any wallet that has made a transaction on Gully Chain or Arbitrum Sepolia\. Brand-new wallet\? Sign in with Gully ID below\./);

  // the sign-in: under the address and the claim button, marked Optional, outlined (not the primary button)
  const gid = $(t, 'gid-optional');
  const btn = $(t, 'gid-signin');
  assert.ok(gid && btn, 'the optional sign-in is on the page');
  assert.ok(before(input, btn) && before(claim, btn), 'address and claim come before the sign-in');
  assert.match($(t, 'gid-optional-pill').textContent, /Optional/);
  assert.ok(btn.classList.contains('sec'), 'secondary (outlined) button');
  assert.match(btn.textContent, /Sign in with Gully ID/);
  assert.equal(t.querySelectorAll('[data-testid="gid-signin"]').length, 1, 'one sign-in on the page');
});

test('optional mode: nothing opens Gully ID by itself and the page never sits on "Opening Gully ID…"', async () => {
  const t = await page();
  await settle(40);
  assert.equal(sdkCalls.signIn, 0, 'no sign-in window without a press');
  const btn = $(t, 'gid-signin');
  assert.equal(btn.disabled, false);
  assert.doesNotMatch(t.textContent, /Opening Gully ID/);
  // typing and claiming never touch Gully ID either
  await typeInto($(t, 'address'), OTHER);
  await submit(t);
  assert.equal(sdkCalls.signIn, 0);
  assert.doesNotMatch(t.textContent, /Opening Gully ID/);
});

test('optional mode, signed out: a pasted address is claimed without any session', async () => {
  const t = await page();
  const steps = () => [...$(t, 'stepper').querySelectorAll('li')].map((li) => (li.classList.contains('done') ? 'done' : li.classList.contains('active') ? 'active' : ''));
  assert.deepEqual(steps(), ['active', '', '']);
  await typeInto($(t, 'address'), OTHER);
  assert.deepEqual(steps(), ['done', 'active', ''], 'an address in the field moves the stepper to "Request ETH"');
  await submit(t);
  assert.deepEqual(claims, [{ address: OTHER, auth: null }], 'one claim, no Authorization header');
  assert.ok($(t, 'done'), 'the done panel shows');
  assert.deepEqual(steps(), ['done', 'done', 'done']);
});

test('optional mode, signed out, brand-new address: the refusal says how to fund it and the sign-in becomes the primary button', async () => {
  const t = await page();
  claimAnswer = () => ({ status: 403, body: { msg: FRESH_MSG } });
  await typeInto($(t, 'address'), FRESH);
  assert.ok($(t, 'gid-signin').classList.contains('sec'));
  await submit(t);
  assert.equal(claims.length, 1);
  const err = t.querySelector('#address-err');
  assert.ok(err, 'the refusal is shown under the field');
  assert.match(err.textContent, /Sign in with Gully ID to fund a brand-new wallet, or use a wallet that has made at least one transaction on Gully Chain or Arbitrum Sepolia\./);
  assert.equal($(t, 'done'), null);
  assert.equal($(t, 'gid-signin').classList.contains('sec'), false, 'the sign-in is now the primary button');
  assert.match(t.querySelector('.toasts').textContent, /Brand-new wallet/);

  // signing in clears the refusal and the same address then goes out with the session
  claimAnswer = () => ({ status: 200, body: { msg: `Txhash: 0x${'cd'.repeat(32)}` } });
  signedIn = session();
  onChange(signedIn);
  await settle(5);
  assert.equal(t.querySelector('#address-err'), null, 'the refusal is gone after signing in');
  assert.equal($(t, 'address').value, FRESH, 'signing in does not replace the address the person typed');
  await submit(t);
  assert.deepEqual(claims.at(-1), { address: FRESH, auth: 'Bearer gid_test_token_1234' });
  assert.ok($(t, 'done'));
});

test('optional mode, signed in: the wallet is filled in, the session is sent, and any other address pays on one press', async () => {
  const t = await page({ sess: session() });
  const input = $(t, 'address');
  assert.equal(input.value, MAIN, 'signed in: the Gully wallet is filled in');
  assert.match($(t, 'gid-who').textContent, /Signed in as owner/);
  assert.match($(t, 'gid-who').textContent, /any address works, brand-new wallets included/);
  assert.match($(t, 'who-is-paid').textContent, /Signed in with Gully ID: any address works, brand-new wallets included\./);
  assert.equal($(t, 'gid-signin'), null);

  await typeInto(input, OTHER);
  const note = $(t, 'not-your-account');
  assert.ok(note, 'a note says this is not the Gully account');
  assert.ok(note.classList.contains('soft'), 'a neutral note, not a warning');
  assert.match(note.textContent, /That is fine: any address works/);
  assert.doesNotMatch($(t, 'claim').textContent, /anyway/);
  await submit(t);
  assert.deepEqual(claims, [{ address: OTHER, auth: 'Bearer gid_test_token_1234' }], 'paid on the first press, with the session');
});

test('optional mode, signed in but the server did not accept the session: the refusal says to sign in again', async () => {
  const t = await page({ sess: session() });
  claimAnswer = () => ({ status: 403, body: { msg: FRESH_MSG } });
  await typeInto($(t, 'address'), FRESH);
  await submit(t);
  assert.match(t.querySelector('#address-err').textContent, /your Gully ID session was not accepted \(it may have expired\)\. Sign out and sign in again/);
});

test('required mode is unchanged: the claim is locked until signed in, then the session is sent', async () => {
  let t = await page({ info: REQUIRED });
  assert.equal($(t, 'address'), null, 'no address field before the sign-in');
  assert.ok($(t, 'claim-why'));
  const btn = $(t, 'gid-signin');
  assert.ok(btn && !btn.classList.contains('sec'), 'the sign-in is the primary button');
  assert.match(t.textContent, /Required: a passkey, no seed phrase\./);
  assert.equal($(t, 'gid-optional'), null);
  assert.deepEqual([...$(t, 'stepper').querySelectorAll('li')].map((li) => li.textContent.replace(/^\d+/, '').trim()), ['Sign in', 'Claim', 'Done']);
  assert.match($(t, 'promise').textContent, /Sign in with Gully ID, then one claim per account, address and IP every 24 hours\./);

  t = await page({ info: REQUIRED, sess: session() });
  assert.equal($(t, 'address').value, MAIN);
  await submit(t);
  assert.deepEqual(claims, [{ address: MAIN, auth: 'Bearer gid_test_token_1234' }]);
});

test('an older server without signin_mode: signin_required false + a client id reads as optional', async () => {
  const { signin_mode, ...old } = OPTIONAL;
  const t = await page({ info: old });
  assert.ok($(t, 'address'));
  assert.ok($(t, 'gid-optional'));
  assert.equal($(t, 'claim').disabled, false);
});

test('no Gully ID at all (mode off): no sign-in anywhere, the address field works', async () => {
  const { gullyid_client, gullyid_issuer, ...off } = { ...OPTIONAL, signin_mode: 'off' };
  const t = await page({ info: off });
  assert.ok($(t, 'address'));
  assert.equal($(t, 'gid-signin'), null);
  assert.doesNotMatch($(t, 'promise').textContent, /Gully ID/);
  await typeInto($(t, 'address'), OTHER);
  await submit(t);
  assert.deepEqual(claims, [{ address: OTHER, auth: null }]);
});

// ---- "Opening Gully ID…" never sticks
async function mountSignIn(props) {
  win.GullyID = fakeSdk();
  signedIn = null;
  sdkCalls.signIn = 0;
  const target = win.document.createElement('div');
  win.document.body.appendChild(target);
  mounted.push(mount(GullySignIn, { target, props: { clientId: 'gully-faucet', ...props } }));
  await settle();
  return target;
}

test('a sign-in window that never answers: Cancel puts the button back at once', async () => {
  const t = await mountSignIn({ required: false });
  signInImpl = () => new Promise(() => {}); // a blocked or abandoned popup never settles
  $(t, 'gid-signin').click();
  await settle(5);
  assert.equal(sdkCalls.signIn, 1);
  assert.match($(t, 'gid-signin').textContent, /Opening Gully ID/);
  assert.equal($(t, 'gid-signin').disabled, true);
  $(t, 'gid-cancel').click();
  await settle(5);
  assert.match($(t, 'gid-signin').textContent, /Sign in with Gully ID/);
  assert.equal($(t, 'gid-signin').disabled, false);
  assert.equal($(t, 'gid-error'), null);
});

test('a sign-in window that never answers: the watchdog puts the button back and says sign-in is optional', async () => {
  const t = await mountSignIn({ required: false, busyMs: 1500 });
  signInImpl = () => new Promise(() => {});
  $(t, 'gid-signin').click();
  flushSync();
  assert.match($(t, 'gid-signin').textContent, /Opening Gully ID/);
  for (let i = 0; i < 100 && $(t, 'gid-signin').disabled; i++) await settle(10);
  assert.match($(t, 'gid-signin').textContent, /Sign in with Gully ID/);
  assert.equal($(t, 'gid-signin').disabled, false);
  assert.match($(t, 'gid-error').textContent, /Gully ID did not answer\. Sign-in is optional: you can paste any address above\./);
});

test('a sign-in that finishes after the watchdog still signs the person in', async () => {
  let finish;
  const filled = [];
  const t = await mountSignIn({ required: false, busyMs: 300, onwallet: (a, u, s) => filled.push([a, s?.accessToken || null]) });
  signInImpl = () => new Promise((r) => { finish = r; });
  $(t, 'gid-signin').click();
  flushSync();
  assert.equal($(t, 'gid-signin').disabled, true);
  for (let i = 0; i < 100 && $(t, 'gid-signin').disabled; i++) await settle(10);
  assert.equal($(t, 'gid-signin').disabled, false, 'watchdog fired');
  finish(session());
  await settle(5);
  assert.deepEqual(filled.at(-1), [MAIN, 'gid_test_token_1234']);
  assert.match($(t, 'gid-who').textContent, /Signed in as owner/);
});
