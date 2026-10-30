// Vendors Gully Chain Night (GN), the shared black design of the Gully Chain builder surfaces, from gully-ops, byte for
// byte: no value is copied by hand. Source of truth: gully-labs/gully-ops sites/_shared/gullychain-night (README.md).
//
//   node scripts/sync-gullychain-night.mjs [--ops ../../gully-ops] [--ref <commit>]     write the files + night.lock.json
//   node scripts/sync-gullychain-night.mjs --check                                       re-hash them against the lock
//
// Writes src/night/{tokens.css,components.css,charts.js} (imported by src/main.js) and public/assets/fonts/
// {inter-var.woff2,Inter-OFL.txt} (GN 2.0 = brand kit v4: tokens.css loads /assets/fonts/inter-var.woff2). Never edit
// those files here: change them in gully-ops and re-run this script with the new commit.
import { execFileSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const WEB = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const LOCK = join(WEB, 'night.lock.json');
const DEFAULT_REF = '8a81bed4'; // gully-ops feat/gc-brand-v4: "GullyChain brand kit v4: GN 2.0 tokens"
export const FILES = [
  ['sites/_shared/gullychain-night/tokens.css', 'src/night/tokens.css'],
  ['sites/_shared/gullychain-night/components.css', 'src/night/components.css'],
  ['sites/_shared/gullychain-night/charts.js', 'src/night/charts.js'],
  ['sites/_shared/brand/fonts/inter-var.woff2', 'public/assets/fonts/inter-var.woff2'],
  ['sites/_shared/brand/fonts/Inter-OFL.txt', 'public/assets/fonts/Inter-OFL.txt'],
];
const sha = (buf) => createHash('sha256').update(buf).digest('hex');

export function check() {
  const lock = JSON.parse(readFileSync(LOCK, 'utf8'));
  const bad = [];
  for (const [, dst] of FILES) {
    const want = lock.files[dst];
    const got = sha(readFileSync(join(WEB, dst)));
    if (want !== got) bad.push(`${dst}: ${got} != lock ${want}`);
  }
  return { lock, bad };
}

const args = process.argv.slice(2);
const arg = (k, d) => { const i = args.indexOf(k); return i >= 0 ? args[i + 1] : d; };
if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  if (args.includes('--check')) {
    const { lock, bad } = check();
    if (bad.length) { console.error(`GN files differ from gully-ops ${lock.ref}:\n  ${bad.join('\n  ')}`); process.exit(1); }
    console.log(`GN files match gully-ops ${lock.ref} (${FILES.length} files)`);
  } else {
    const ops = resolve(arg('--ops', join(WEB, '..', '..', 'gully-ops')));
    const ref = execFileSync('git', ['-C', ops, 'rev-parse', arg('--ref', DEFAULT_REF)], { encoding: 'utf8' }).trim();
    const files = {};
    for (const [src, dst] of FILES) {
      const buf = execFileSync('git', ['-C', ops, 'show', `${ref}:${src}`], { maxBuffer: 1 << 26 });
      mkdirSync(dirname(join(WEB, dst)), { recursive: true });
      writeFileSync(join(WEB, dst), buf);
      files[dst] = sha(buf);
    }
    writeFileSync(LOCK, JSON.stringify({ source: 'gully-labs/gully-ops', ref, paths: Object.fromEntries(FILES.map(([s, d]) => [d, s])), files }, null, 2) + '\n');
    console.log(`wrote ${FILES.length} GN files from gully-ops ${ref}`);
  }
}
