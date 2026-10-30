// node --test tests/: owner rule 2026-10-30 ("need icons on the whole stack"). Every link in the header nav, the mobile
// menu, the footer and the 404 page nav carries a Lucide icon (20 px, stroke 1.75, muted, mint when active), label kept.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const WEB = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const read = (p) => readFileSync(join(WEB, p), 'utf8');
const faucet = read('src/Faucet.svelte');
const markup = faucet.slice(faucet.indexOf('</script>', faucet.indexOf('const icon = ')), faucet.indexOf('<style>'));

// the Lucide paths the faucet ships (icon() falls back to "info" for an unknown name, so check the names are real)
const ICONS = new Set([...faucet.slice(faucet.indexOf('const ICONS = {'), faucet.indexOf('};', faucet.indexOf('const ICONS = {'))).matchAll(/^\s+([a-z]+): '/gm)].map((m) => m[1]));
const list = (name) => {
  const start = faucet.indexOf(`const ${name} = [`);
  assert.ok(start >= 0, name);
  return [...faucet.slice(start, faucet.indexOf('];', start)).matchAll(/\{ name: '([^']+)'([^}]*)\}/g)].map((m) => ({ name: m[1], icon: m[2].match(/icon: '([a-z]+)'/)?.[1] }));
};
const links = (block) => [...block.matchAll(/<a\b(?:[^>{]|\{[^}]*\})*>([\s\S]*?)<\/a>/g)].map((m) => m[1]);

test('every header nav, mobile menu and footer list entry names a Lucide icon the faucet ships', () => {
  for (const name of ['NAV', 'PRODUCTS', 'SOCIALS']) {
    const items = list(name);
    assert.ok(items.length >= 4, name);
    for (const it of items) {
      assert.ok(it.icon, `${name} ${it.name} has no icon`);
      assert.ok(ICONS.has(it.icon), `${name} ${it.name}: unknown icon ${it.icon}`);
    }
  }
  for (const n of ['menu', 'x', 'branch', 'layers']) assert.ok(ICONS.has(n), n);
});

test('every link in every nav, the mobile menu and the footer link lists renders an icon next to its label', () => {
  const blocks = [
    ...[...markup.matchAll(/<nav\b[\s\S]*?<\/nav>/g)].map((m) => m[0]),
    ...[...markup.matchAll(/<ul class="fnav"[\s\S]*?<\/ul>/g)].map((m) => m[0]),
  ];
  const labels = blocks.map((b) => b.match(/aria-label="([^"]+)"/)?.[1]);
  for (const l of ['Gully Chain', 'Gully Chain menu', 'GullyChain products', 'Official GullyChain channels']) assert.ok(labels.includes(l), `missing nav group ${l}`);
  for (const b of blocks) {
    const as = links(b);
    assert.ok(as.length > 0, b.slice(0, 80));
    for (const a of as) {
      assert.match(a, /^<span class="ic">\{@html icon\([^)]+\)\}<\/span>/, `link without a leading icon: ${a}`);
      assert.match(a.replace(/<span class="ic">[\s\S]*?<\/span>/, ''), /[A-Za-z]/, `link without a text label: ${a}`);
    }
  }
  // the mobile menu exists below 900 px and lists the same NAV entries
  assert.match(markup, /<button type="button" class="menubtn" aria-label="Menu" aria-expanded=\{menuOpen\} aria-controls="mnav"/);
  assert.match(markup, /<nav class="mnav" id="mnav" aria-label="Gully Chain menu"><ul>\{#each NAV as l\}/);
});

test('nav icons are 20 px, stroke 1.75, muted by default and mint when active or hovered', () => {
  const css = faucet.slice(faucet.indexOf('<style>'));
  assert.match(css, /\.ic :global\(svg\) \{[^}]*stroke-width: 1\.75/);
  assert.match(css, /\.topnav a \.ic, \.mnav a \.ic, footer \.fnav a \.ic \{ width: 20px; height: 20px; color: var\(--gn-text-3\); \}/);
  assert.match(css, /\.topnav a\[aria-current="page"\] \.ic, \.mnav a\[aria-current="page"\] \.ic, footer \.fnav a\[aria-current="page"\] \.ic \{ color: var\(--gn-link\); \}/);
});

test('404 page: every nav link has an inline Lucide svg, muted and mint on hover', () => {
  const h = read('404.html');
  const nav = h.match(/<ul class="nf-links"[\s\S]*?<\/ul>/)[0];
  const as = links(nav);
  assert.ok(as.length >= 6);
  for (const a of as) assert.match(a, /^<svg class="gn-ic" viewBox="0 0 24 24" aria-hidden="true">[\s\S]+<\/svg> [A-Z]/, a);
  const css = read('src/notfound.css');
  assert.match(css, /\.nf-links a \.gn-ic \{ color: var\(--gn-text-3\); \}/);
  assert.match(css, /\.nf-links a:hover \.gn-ic, \.nf-links a\[aria-current="page"\] \.gn-ic \{ color: var\(--gn-link\); \}/);
  assert.match(read('src/night/components.css'), /\.gn-ic \{ width: 20px; height: 20px;[^}]*stroke-width: 1\.75/);
});
