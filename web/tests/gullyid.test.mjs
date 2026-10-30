// node --test tests/: the Gully ID SDK loader (v1.js first, legacy sdk.js fallback) and the "Sign in with Gully ID"
// component rendering the official SDK button (GullySignIn.svelte compiled with svelte/compiler, mounted in happy-dom).
import { test, beforeEach } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, writeFileSync, mkdtempSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { Window } from 'happy-dom';
import { compile } from 'svelte/compiler';

const WEB = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const win = new Window({ url: 'https://faucet.gullychain.xyz/' });
for (const k of ['window', 'document', 'navigator', 'HTMLElement', 'Element', 'Node', 'Text', 'Comment', 'DocumentFragment', 'MutationObserver', 'getComputedStyle', 'requestAnimationFrame']) {
  if (!(k in globalThis) || k === 'window' || k === 'document') Object.defineProperty(globalThis, k, { value: k === 'window' ? win : win[k], configurable: true, writable: true });
}

// fake <script> loading: the test decides per URL whether it loads (and defines GullyID) or errors
let behaviour = () => 'error';
const loaded = [];
const realAppend = win.document.head.appendChild.bind(win.document.head);
win.document.head.appendChild = (el) => {
  if (el.tagName !== 'SCRIPT') return realAppend(el);
  loaded.push(el.src);
  queueMicrotask(() => {
    const r = behaviour(el.src);
    if (r === 'error') return el.onerror?.();
    if (r === 'sdk') win.GullyID = fakeSdk();
    el.onload?.();
  });
  return el;
};

const renders = [];
function fakeSdk() {
  return {
    version: '1.2.0',
    init: (o) => ({
      opts: o,
      handleRedirect: async () => null,
      session: () => null,
      onChange: () => () => {},
      renderButton(el, opts) {
        renders.push(opts);
        const b = win.document.createElement('button');
        b.textContent = 'Sign in with Gully ID';
        el.appendChild(b);
        return () => b.remove();
      },
    }),
  };
}

const { loadGullyID, sdkUrls } = await import('../src/gullyid.js');

beforeEach(() => {
  delete win.GullyID;
  loaded.length = 0;
  renders.length = 0;
});

test('SDK URLs: the 1.x channel first, then the legacy /sdk.js, on gullyid.xyz', () => {
  assert.deepEqual(sdkUrls(), ['https://gullyid.xyz/sdk/v1.js', 'https://gullyid.xyz/sdk.js']);
  assert.deepEqual(sdkUrls('https://gullyid.xyz/'), ['https://gullyid.xyz/sdk/v1.js', 'https://gullyid.xyz/sdk.js']);
});

test('loads /sdk/v1.js when it is live', async () => {
  behaviour = () => 'sdk';
  const g = await loadGullyID();
  assert.equal(typeof g.init, 'function');
  assert.deepEqual(loaded, ['https://gullyid.xyz/sdk/v1.js']);
});

test('falls back to /sdk.js while /sdk/v1.js is 404', async () => {
  behaviour = (src) => (src.endsWith('/sdk/v1.js') ? 'error' : 'sdk');
  const g = await loadGullyID();
  assert.equal(typeof g.init, 'function');
  assert.deepEqual(loaded, ['https://gullyid.xyz/sdk/v1.js', 'https://gullyid.xyz/sdk.js']);
});

test('rejects (button hidden) when neither URL loads', async () => {
  behaviour = () => 'error';
  await assert.rejects(loadGullyID());
});

test('GullySignIn renders the official Sign in with Gully ID button', async () => {
  behaviour = () => 'sdk';
  const src = readFileSync(join(WEB, 'src/GullySignIn.svelte'), 'utf8');
  const { js } = compile(src, { generate: 'client', filename: 'GullySignIn.svelte', css: 'injected' });
  const dir = mkdtempSync(join(WEB, 'node_modules', '.gid-test-'));
  writeFileSync(join(dir, 'gullyid.js'), readFileSync(join(WEB, 'src/gullyid.js')));
  writeFileSync(join(dir, 'GullySignIn.js'), js.code);
  const { mount, flushSync } = await import('svelte');
  const { default: GullySignIn } = await import(pathToFileURL(join(dir, 'GullySignIn.js')).href);
  const target = win.document.createElement('div');
  win.document.body.appendChild(target);
  mount(GullySignIn, { target, props: { clientId: 'gully-faucet', required: true } });
  for (let i = 0; i < 20 && !target.querySelector('button'); i++) {
    await new Promise((r) => setTimeout(r, 5));
    flushSync();
  }
  assert.equal(target.querySelector('button')?.textContent, 'Sign in with Gully ID');
  assert.equal(renders[0].theme, 'dark');
  assert.ok(target.querySelector('.gid.on'), 'the sign-in row is shown');
});
