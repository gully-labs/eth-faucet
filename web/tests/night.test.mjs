// node --test tests/: the vendored Gully Chain Night files and brand-kit logos are byte-identical to their sources.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { check } from '../scripts/sync-gullychain-night.mjs';

const WEB = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const sha = (p) => createHash('sha256').update(readFileSync(join(WEB, p))).digest('hex');
const src = () => readFileSync(join(WEB, 'src/Faucet.svelte'), 'utf8') + readFileSync(join(WEB, 'src/App.svelte'), 'utf8');

test('GN tokens/components/charts/fonts equal the gully-ops commit in night.lock.json', () => {
  const { bad } = check();
  assert.deepEqual(bad, []);
});

test('the entry imports the shared GN tokens and components', () => {
  const main = readFileSync(join(WEB, 'src/main.js'), 'utf8');
  assert.match(main, /import '\.\/night\/tokens\.css'/);
  assert.match(main, /import '\.\/night\/components\.css'/);
});

// SHA-256 of the brand kit v4 files (gully-ops sites/_shared/brand/kit-v4 = Desktop\GULLY LABS GRAPHICS\Gully Chain\
// brand-kit-v4), the landing favicons, the faucet OG card (sites/_shared/brand/og), gc-switcher 1.8.2, gc-connect 1.0.0 and the Gully Labs
// dark badge (gully-ops sites/_shared/brand); recorded 2026-10-30 from gully-ops feat/gc-brand-v4 8a81bed4.
const KIT = JSON.parse(readFileSync(join(WEB, 'tests/kit-sha256.json'), 'utf8'));
test('brand-kit logos and badges are byte-identical to the kit', () => {
  for (const [p, h] of Object.entries(KIT)) assert.equal(sha(p), h, p);
});

test('dark pages use only the dark-background logo variants', () => {
  const s = src();
  assert.doesNotMatch(s, /for-light-bg|powered-by-light|gully-labs-light/);
  assert.match(s, /src="\/brand\/logo\/gc-icon\.svg" width="32" height="32"/);
  assert.match(s, /src="\/brand\/badges\/badge-dark\.png" width="180" height="48"/);
  assert.match(s, /src="\/brand\/mascots\/mascot-builder\.png" width="630" height="502"/);
  // kit v4 has no wordmark file: no v3 wordmark, 3D logo or v3 badge anywhere
  assert.doesNotMatch(s, /gullychain-wordmark|gullychain-logo-transparent|gullychain-powered-by|gullychain-mascot/);
  assert.match(s, /powered-by-gully-labs-dark/);
  assert.match(s, /<gc-switcher [^>]*theme="dark"/);
});

test('no Studio Light colour literals are left in the styles', () => {
  const s = src();
  for (const lit of ['#ffffff', '#fff;', '#0b0d12', '#0a86c2', '#075f8c', '#f6f7f9', '#e6e8ec', 'rgba(255, 255, 255, 0.92)']) {
    assert.ok(!s.toLowerCase().includes(lit), `literal ${lit} still in the faucet styles`);
  }
  assert.doesNotMatch(s, /family=Inter/);
});

test('theme-color and color-scheme are the GN 2.0 kit ink', () => {
  const html = readFileSync(join(WEB, 'index.html'), 'utf8');
  assert.match(html, /name="theme-color" content="#0A0F0E"/);
  assert.match(html, /name="color-scheme" content="dark"/);
  assert.doesNotMatch(html, /fonts\.googleapis/);
  const mf = JSON.parse(readFileSync(join(WEB, 'public/site.webmanifest'), 'utf8'));
  assert.equal(mf.theme_color.toUpperCase(), '#0A0F0E');
  assert.equal(mf.background_color.toUpperCase(), '#0A0F0E');
  assert.match(html, /<link rel="preload" href="\/assets\/fonts\/inter-var\.woff2"/);
  assert.match(html, /<meta property="og:image" content="https:\/\/faucet\.gullychain\.xyz\/og-image\.png"/);
  assert.match(html, /<meta name="twitter:image" content="https:\/\/faucet\.gullychain\.xyz\/og-image\.png"/);
});

// Design-gap audit 2026-10-29: one click to every key Gully Chain surface, switcher 1.7.0, branded 404, no retired token.
test('header nav reaches dev, docs, bridge, explorer, ecosystem and status, each with an icon', () => {
  const s = readFileSync(join(WEB, 'src/Faucet.svelte'), 'utf8');
  const nav = s.slice(s.indexOf('const NAV = ['), s.indexOf('];', s.indexOf('const NAV = [')));
  for (const u of ['https://dev.gullychain.xyz/', 'https://docs.gullychain.xyz/', 'https://testnet-bridge.gullychain.xyz/', "CHAIN.explorer + '/'", 'https://gullychain.xyz/ecosystem/', 'https://status.gullylabs.xyz/']) assert.ok(nav.includes(u), u);
  assert.equal((nav.match(/icon: '/g) || []).length, 6, 'every nav item has an icon');
  assert.match(s, /<nav class="topnav"[\s\S]*?\{@html icon\(l\.icon\)\}<\/span><span class="tl">\{l\.name\}/);
  assert.match(s.slice(s.indexOf('const PRODUCTS = [')), /\{ name: 'Ecosystem', href: 'https:\/\/gullychain\.xyz\/ecosystem\/', icon: 'grid' \}/, 'footer links the ecosystem page');
});

test('gc-switcher 1.8.2 (kit v4 dark theme, Gully Chain products only), loaded with its version on both pages', () => {
  const js = readFileSync(join(WEB, 'public/gc-switcher.js'), 'utf8');
  assert.match(js, /GC_SWITCHER_VERSION = '1\.8\.2'/);
  assert.doesNotMatch(js.slice(js.indexOf('var PRODUCTS')), /href: 'https:\/\/[a-z.]*kre8\.fun/, '1.8.2: no kre8 entries in the grid');
  for (const f of ['index.html', '404.html']) assert.match(readFileSync(join(WEB, f), 'utf8'), /gc-switcher\.js\?v=1\.8\.2/, f);
});

// GN 2.0 / brand kit v4: no kit v3 colour literal or Exo 2 left in the faucet's own sources (src/night/ is vendored).
test('no kit v3 colour literals or Exo 2 in the faucet sources', () => {
  const files = ['src/App.svelte', 'src/Faucet.svelte', 'src/GullySignIn.svelte', 'src/notfound.css', 'index.html', '404.html', 'public/site.webmanifest'];
  const v3 = /#(05080D|0A1018|0F1722|2563FF|00B2FA|46FB04|F2F5F7|8A97A8|C5CDD7|6E93FF)(?![0-9a-f])|rgba?\(\s*0\s*,\s*178\s*,\s*250|Exo 2|exo2/i;
  for (const f of files) assert.doesNotMatch(readFileSync(join(WEB, f), 'utf8'), v3, f);
});

test('the page does not advertise the retired glUSD drip', () => {
  const s = readFileSync(join(WEB, 'src/Faucet.svelte'), 'utf8');
  assert.match(s, /RETIRED_TOKENS = \['glUSD'\]/);
  assert.match(s, /\{#if faucetInfo\.token && !RETIRED_TOKENS\.includes\(faucetInfo\.token\.symbol\)\}Each claim also sends/);
});

test('404 page: noindex, kit dark logo, switcher, links to every key surface, GN tokens only', () => {
  const h = readFileSync(join(WEB, '404.html'), 'utf8');
  assert.match(h, /name="robots" content="noindex"/);
  assert.match(h, /src="\/brand\/logo\/gc-icon\.svg" width="32" height="32"/);
  assert.match(h, /name="theme-color" content="#0A0F0E"/);
  assert.match(h, /<gc-switcher [^>]*theme="dark"/);
  for (const u of ['dev.gullychain.xyz', 'docs.gullychain.xyz', 'testnet-bridge.gullychain.xyz', 'testnet-explorer.gullychain.xyz', 'gullychain.xyz/ecosystem/', 'status.gullylabs.xyz']) assert.ok(h.includes(`https://${u}`), u);
  assert.doesNotMatch(readFileSync(join(WEB, 'src/notfound.css'), 'utf8'), /#[0-9a-f]{3,6}\b|rgba?\(/i);
});
