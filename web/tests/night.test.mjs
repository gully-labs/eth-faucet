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

// SHA-256 of the brand kit v3 files (C:\Users\Admin\Desktop\GULLY LABS GRAPHICS\Gully Chain\brand-kit) and the Gully
// Labs dark badge (gully-ops sites/_shared/brand); recorded 2026-10-29.
const KIT = JSON.parse(readFileSync(join(WEB, 'tests/kit-sha256.json'), 'utf8'));
test('brand-kit logos and badges are byte-identical to the kit', () => {
  for (const [p, h] of Object.entries(KIT)) assert.equal(sha(p), h, p);
});

test('dark pages use only the dark-background logo variants', () => {
  const s = src();
  assert.doesNotMatch(s, /for-light-bg|powered-by-light|gully-labs-light/);
  assert.match(s, /gullychain-logo-transparent\.png/);
  assert.match(s, /gullychain-wordmark-for-dark-bg\.svg/);
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

test('theme-color and color-scheme are the GN night', () => {
  const html = readFileSync(join(WEB, 'index.html'), 'utf8');
  assert.match(html, /name="theme-color" content="#05080D"/);
  assert.match(html, /name="color-scheme" content="dark"/);
  assert.doesNotMatch(html, /fonts\.googleapis/);
  const mf = JSON.parse(readFileSync(join(WEB, 'public/site.webmanifest'), 'utf8'));
  assert.equal(mf.theme_color.toUpperCase(), '#05080D');
  assert.equal(mf.background_color.toUpperCase(), '#05080D');
});
