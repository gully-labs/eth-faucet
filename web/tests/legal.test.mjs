// node --test tests/: Gully Chain's one Terms of Service and Privacy Policy (gullychain.xyz) are linked from every
// faucet footer (app + 404) in a new tab, and the claim form says that requesting testnet ETH means agreeing to them.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, readdirSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const WEB = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const read = (p) => readFileSync(join(WEB, p), 'utf8');
const faucet = read('src/Faucet.svelte');
const notFound = read('404.html');
const TERMS = 'https://gullychain.xyz/terms';
const PRIVACY = 'https://gullychain.xyz/privacy';

test('the app footer lists Terms and Privacy on gullychain.xyz, new tab', () => {
  assert.match(faucet, new RegExp(`const TERMS = '${TERMS}';`));
  assert.match(faucet, new RegExp(`const PRIVACY = '${PRIVACY}';`));
  assert.match(faucet, /\{ name: 'Terms', href: TERMS, icon: 'file' \}/);
  assert.match(faucet, /\{ name: 'Privacy', href: PRIVACY, icon: 'shield' \}/);
  const foot = faucet.slice(faucet.indexOf('<footer>'), faucet.indexOf('</footer>'));
  assert.match(foot, /<ul class="fnav" aria-label="Legal" data-testid="legal">\{#each LEGAL as l\}<li><a href=\{l\.href\} target="_blank" rel="noopener noreferrer">/);
  assert.match(faucet, /file: '<path /, 'the file icon exists');
});

test('the 404 page footer links Terms and Privacy on gullychain.xyz, new tab', () => {
  for (const u of [TERMS, PRIVACY]) assert.ok(notFound.includes(`<a href="${u}" target="_blank" rel="noopener noreferrer">`), u);
  assert.ok(notFound.indexOf('data-testid="legal"') < notFound.indexOf('</main>'));
});

test('the consent line sits next to the claim button and links both pages', () => {
  const form = faucet.slice(faucet.indexOf('<form id="claim"'), faucet.indexOf('</form>'));
  const btn = form.indexOf('data-testid="claim"'), line = form.indexOf('data-testid="consent"');
  assert.ok(btn > 0 && line > btn, `claim ${btn}, consent ${line}`);
  const p = form.slice(line, form.indexOf('</p>', line));
  assert.match(p, /By requesting testnet ETH you agree to the Gully Chain <a href=\{TERMS\} target="_blank" rel="noopener noreferrer">Terms/);
  assert.match(p, /and <a href=\{PRIVACY\} target="_blank" rel="noopener noreferrer">Privacy Policy/);
  assert.match(p, /Testnet ETH has no value\./);
});

// The Go 301s for /terms and /privacy are covered by internal/server/legal_test.go (the web tests run with web/ only).

test('no draft legal marker and no old docs.gullychain.xyz/privacy link in the web sources', () => {
  const files = ['index.html', '404.html', ...readdirSync(join(WEB, 'src'), { recursive: true }).map((f) => `src/${f}`)].filter((f) => /\.(svelte|js|css|html)$/.test(f));
  for (const f of files) {
    const t = read(f);
    assert.doesNotMatch(t, /pending legal review|review by counsel|legal draft/i, f);
    assert.ok(!t.includes('docs.gullychain.xyz/privacy'), f);
  }
});
